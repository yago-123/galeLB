#ifndef COMMON_H
#define COMMON_H

// Ensure __u32 and __u16 are defined
#include <linux/types.h>

typedef struct {
    __u32 ip;   // 4 bytes
    __u16 port; // 2 bytes
    __u16 pad;  // 2 bytes (padding for alignment)
} ip_port_key;

#endif // COMMON_H
