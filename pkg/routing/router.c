// +build ignore
#include <linux/bpf.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/tcp.h>

#include <bpf/bpf_helpers.h>


#define BACKEND_IP 0xC0A80070  // 192.168.0.112 (fixed Backend)
#define LB_IP 0xC0A800F2       // 192.168.0.242 (fixed Load Balancer)

// represent protocol number for TCP
#define IPPROTO_TCP 6

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 65536);
    __type(key, struct iphdr);
    __type(value, __u32);
} snat_map SEC(".maps");

SEC("tc_ingress")  // Attach this to eth0 (incoming traffic)
int dnat_prog(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end) return TC_ACT_OK;

    if (eth->h_proto != __constant_htons(ETH_P_IP)) return TC_ACT_OK;

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end) return TC_ACT_OK;

    if (ip->protocol != IPPROTO_TCP) return TC_ACT_OK;

    struct tcphdr *tcp = (void *)(ip + 1);
    if ((void *)(tcp + 1) > data_end) return TC_ACT_OK;

    // Convert from network byte order and check if the dest port is 8080
    if (tcp->dest != __constant_htons(8080)) {
        return TC_ACT_OK;  // Ignore if not port 8080
    }

    // Keep track of the original destination and source address
    __u32 old_daddr = ip->daddr;
    __u32 old_saddr = ip->saddr;

    // Modify the destination address to the backend IP
    ip->daddr = __constant_htonl(BACKEND_IP);
    // Modify the source address to the load balancer to prevent breaking the connection
    ip->saddr = __constant_htonl(LB_IP);

    // Update the checksum for both IP and TCP headers
    bpf_l3_csum_replace(skb, offsetof(struct iphdr, check), old_daddr, ip->daddr, sizeof(__u32));
    bpf_l4_csum_replace(skb, offsetof(struct tcphdr, check), old_daddr, ip->daddr, sizeof(__u32));

    bpf_l3_csum_replace(skb, offsetof(struct iphdr, check), old_saddr, ip->saddr, sizeof(__u32));
    bpf_l4_csum_replace(skb, offsetof(struct tcphdr, check), old_saddr, ip->saddr, sizeof(__u32));

    return TC_ACT_OK;
}

SEC("tc_egress")  // Attach this to eth1 (forwarding to backend)
int snat_prog(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end) return TC_ACT_OK;

    if (eth->h_proto != __constant_htons(ETH_P_IP)) return TC_ACT_OK;

    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end) return TC_ACT_OK;

    if (ip->protocol != IPPROTO_TCP) return TC_ACT_OK;

    struct tcphdr *tcp = (void *)(ip + 1);
    if ((void *)(tcp + 1) > data_end) return TC_ACT_OK;

    struct iphdr key = *ip;
    __u32 new_saddr = __constant_htonl(LB_IP); // SNAT to LB IP

    __u32 *orig_saddr = bpf_map_lookup_elem(&snat_map, &key);
    if (orig_saddr) {
        ip->saddr = *orig_saddr;
    } else {
        ip->saddr = new_saddr;
        bpf_map_update_elem(&snat_map, &key, &ip->saddr, BPF_ANY);
    }

    bpf_l3_csum_replace(skb, offsetof(struct iphdr, check), key.saddr, ip->saddr, sizeof(__u32));
    bpf_l4_csum_replace(skb, offsetof(struct tcphdr, check), key.saddr, ip->saddr, sizeof(__u32));

    return TC_ACT_OK;
}

char _license[] SEC("license") = "GPL";
