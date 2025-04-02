// +build ignore

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/tcp.h>
#include <linux/if_packet.h>
#include <linux/ptrace.h>
#include <linux/bpf_common.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#include "constants.h"

// defined in pkg/common/
#include "common.h"

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, MAX_NUMBER_VIRTUAL_NODE_ENTRIES);
    __type(key, __u32);
    __type(value, __u32);
} xdp_array_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_NUMBER_VIRTUAL_NODE_ENTRIES);
    __type(key, __u32);
    __type(value, ip_port_key);
} xdp_hash_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, __u32);  // Source IP
    __type(value, __u32); // NAT-ed IP
} dnat_map SEC(".maps");

SEC("tc")
int dnat_prog(struct __sk_buff *skb) {
    // Ensure safe access to socket buffer data
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;

    // Verify validity of ethernet bounds
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end) return TC_ACT_OK;

    // Verify that it's IPV4 packet
    // todo(): add support for IPv6
    if (eth->h_proto != __constant_htons(ETH_P_IP)) return TC_ACT_OK;

    // Verify validity of IP header
    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end) return TC_ACT_OK;

    // Verify that it's TCP packet
    // todo(): add support for UDP
    if (ip->protocol != IPPROTO_TCP) return TC_ACT_OK;

    // Verify validity of TCP header
    struct tcphdr *tcp = (void *)(ip + 1);
    if ((void *)(tcp + 1) > data_end) return TC_ACT_OK;

    // Check if DNAT exists for the source IP
    __u32 *new_daddr = bpf_map_lookup_elem(&dnat_map, &ip->daddr);
    if (!new_daddr) return TC_ACT_OK;

    // Modify destination IP
    __u32 old_daddr = ip->daddr;
    ip->daddr = *new_daddr;

    // Update checksums to account for the new destination IP
    bpf_l3_csum_replace(skb, offsetof(struct iphdr, check), old_daddr, *new_daddr, sizeof(__u32));
    bpf_l4_csum_replace(skb, offsetof(struct tcphdr, check), old_daddr, *new_daddr, sizeof(__u32));

    return TC_ACT_OK;
}

SEC("license") char _license[] = "GPL";
