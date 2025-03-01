// +build ignore

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/if_packet.h>
#include <linux/ptrace.h>
#include <linux/bpf_common.h>
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

SEC("xdp")
int xdp_router(struct __sk_buff *skb) {
    // Drop the packet
    return XDP_PASS;
}

SEC("license") char _license[] = "GPL";
