#include <linux/types.h>
int attach_bpf_prog_to_if(char *, __u32, char *);
int do_detach(int, int);
int do_attach(int, int, __u32);

struct input_args
{
    __u32 xdp_flags;
    char *ifname;
    char *filename;
};