// +build ignore

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/if_packet.h>
#include <linux/ptrace.h>
#include <linux/bpf_common.h>

#define SEC(NAME) __attribute__((section(NAME), used))

SEC("xdp")
int xdp_router(struct __sk_buff *skb) {
    // Drop the packet
    return XDP_DROP;
}

SEC("license") char _license[] = "GPL";