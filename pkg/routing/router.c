// +build ignore
#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <bpf/bpf_helpers.h>

#define BACKEND_IP 0xC0A80070  // 192.168.0.112
#define LB_IP      0xC0A800F2  // 192.168.0.242
#define IPPROTO_TCP 6
#define TARGET_PORT 8080

struct conn_tuple {
    __u32 src_ip;
    __u32 dst_ip;
    __u16 src_port;
    __u16 dst_port;
    __u8  protocol;
};

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 65536);
    __type(key, struct conn_tuple);
    __type(value, __u32); // original source IP
} conntrack_map SEC(".maps");

// parse_headers parses the Ethernet, IP, and TCP headers from the skb
static __always_inline int parse_headers(struct __sk_buff *skb, struct ethhdr **eth, struct iphdr **ip, struct tcphdr **tcp) {
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    *eth = data;
    if ((void *)(*eth + 1) > data_end) return 0;
    if ((*eth)->h_proto != __constant_htons(ETH_P_IP)) return 0;

    *ip = (void *)(*eth + 1);
    if ((void *)(*ip + 1) > data_end) return 0;
    if ((*ip)->protocol != IPPROTO_TCP) return 0;

    *tcp = (void *)(*ip + 1);
    if ((void *)(*tcp + 1) > data_end) return 0;

    return 1;
}

SEC("tc_ingress")
int dnat_prog(struct __sk_buff *skb) {
    struct ethhdr *eth;
    struct iphdr *ip;
    struct tcphdr *tcp;

    if (!parse_headers(skb, &eth, &ip, &tcp))
        return TC_ACT_OK;

    if (tcp->dest != __constant_htons(TARGET_PORT))
        return TC_ACT_OK;

    // Build connection tuple
    struct conn_tuple tuple = {
        .src_ip = ip->saddr,
        .dst_ip = ip->daddr,
        .src_port = tcp->source,
        .dst_port = tcp->dest,
        .protocol = ip->protocol,
    };

    // Save original source IP to keep track of the connection
    __u32 orig_ip = ip->saddr;
    bpf_map_update_elem(&conntrack_map, &tuple, &orig_ip, BPF_ANY);

    // DNAT + SNAT
    __u32 old_daddr = ip->daddr;
    __u32 old_saddr = ip->saddr;

    ip->daddr = __constant_htonl(BACKEND_IP);
    ip->saddr = __constant_htonl(LB_IP);

    // Checksums
    bpf_l3_csum_replace(skb, offsetof(struct iphdr, check), old_daddr, ip->daddr, sizeof(__u32));
    bpf_l4_csum_replace(skb, offsetof(struct tcphdr, check), old_daddr, ip->daddr, sizeof(__u32));

    bpf_l3_csum_replace(skb, offsetof(struct iphdr, check), old_saddr, ip->saddr, sizeof(__u32));
    bpf_l4_csum_replace(skb, offsetof(struct tcphdr, check), old_saddr, ip->saddr, sizeof(__u32));

    return TC_ACT_OK;
}

SEC("tc_egress")
int snat_prog(struct __sk_buff *skb) {
    struct ethhdr *eth;
    struct iphdr *ip;
    struct tcphdr *tcp;

    if (!parse_headers(skb, &eth, &ip, &tcp))
        return TC_ACT_OK;

    // Build reverse tuple to find original client
    struct conn_tuple rev_tuple = {
        .src_ip = ip->daddr,
        .dst_ip = ip->saddr,
        .src_port = tcp->dest,
        .dst_port = tcp->source,
        .protocol = ip->protocol,
    };

    // Check if we have a mapping for the reverse tuple in order to restore the original source IP
    __u32 *orig_ip = bpf_map_lookup_elem(&conntrack_map, &rev_tuple);
    if (!orig_ip)
        return TC_ACT_OK;

    __u32 old_saddr = ip->saddr;
    ip->saddr = *orig_ip;

    bpf_l3_csum_replace(skb, offsetof(struct iphdr, check), old_saddr, ip->saddr, sizeof(__u32));
    bpf_l4_csum_replace(skb, offsetof(struct tcphdr, check), old_saddr, ip->saddr, sizeof(__u32));

    return TC_ACT_OK;
}


char _license[] SEC("license") = "GPL";
