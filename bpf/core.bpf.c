#include "headers/vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include "headers/commons.h"
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>

#define ARGSIZE 128
#define TASK_COMM_LEN 16
#define TOTAL_MAX_ARGS 60
#define FULL_MAX_ARGS_ARR (TOTAL_MAX_ARGS * ARGSIZE)
#define INVALID_UID ((uid_t)-1)
#define BASE_EVENT_SIZE (size_t)(&((struct event *)0)->args)
#define EVENT_SIZE(e) (BASE_EVENT_SIZE + e->args_size)
#define LAST_ARG (FULL_MAX_ARGS_ARR - ARGSIZE)
#define ETH_P_IP 0x0800 /* Internet Protocol packet	*/

#ifndef memcpy
#define memcpy(dest, src, n) __builtin_memcpy((dest), (src), (n))
#endif

// BPF_RINGBUF(events, 1024)

// struct file_event {
//     char filename[64];
//     char args[128];
//     unsigned int args_size;
//     unsigned int args_count;
// };

// typedef struct args {
//     unsigned long args[6];
// } args_t;

// typedef struct task_context {
//     u64 start_time; // thread's start time
//     u64 cgroup_id;
//     u32 pid;       // PID as in the userspace term
//     u32 tid;       // TID as in the userspace term
//     u32 ppid;      // Parent PID as in the userspace term
//     u32 host_pid;  // PID in host pid namespace
//     u32 host_tid;  // TID in host pid namespace
//     u32 host_ppid; // Parent PID in host pid namespace
//     u32 uid;
//     u32 mnt_id;
//     u32 pid_id;
//     char comm[TASK_COMM_LEN];
//     char uts_name[TASK_COMM_LEN];
//     u32 flags;
// } task_context_t;

// typedef struct syscall_data {
//     uint id;           // Current syscall id
//     args_t args;       // Syscall arguments
//     unsigned long ts;  // Timestamp of syscall entry
//     unsigned long ret; // Syscall ret val. May be used by syscall exit tail calls.
// } syscall_data_t;

// typedef struct task_info {
//     task_context_t context;
//     syscall_data_t syscall_data;
//     bool syscall_traced;  // indicates that syscall_data is valid
//     bool recompute_scope; // recompute should_trace (new task/context changed/policy changed)
//     bool new_task;        // set if this task was started after tracee. Used with new_pid filter
//     bool follow;          // set if this task was traced before. Used with the follow filter
//     int should_trace;     // last decision of should_trace()
//     u8 container_state;   // the state of the container the task resides in
// } task_info_t;

// typedef struct event_context {
//     u64 ts; // Timestamp
//     task_context_t task;
//     u32 eventid;
//     u32 padding;
//     s64 retval;
//     u32 stack_id;
//     u16 processor_id; // The ID of the processor which processed the event
//     u8 argnum;
// } event_context_t;

// typedef struct event_data {
//     event_context_t context;
//     struct task_struct *task;
//     void *ctx;
//     task_info_t *task_info;
// } event_data_t;

// static __always_inline int init_context(event_context_t *context, struct task_struct *task) {
//     u64 id = bpf_get_current_pid_tgid();
//     bpf_probe_read_kernel(&context->task.start_time, sizeof(task->start_time), &task->start_time);
//     context->task.host_tid = id;
//     context->task.host_pid = id >> 32;
//     bpf_probe_read_kernel(&context->task.host_ppid, sizeof(task->nsproxy), &task->nsproxy);
//     bpf_probe_read_kernel(&context->task.tid, sizeof(task->thread_pid), &task->thread_pid);
    
//     context->task.uid = bpf_get_current_uid_gid();
//     context->task.flags = 0;
//     bpf_get_current_comm(&context->task.comm, sizeof(context->task.comm));
//     context->ts = bpf_ktime_get_ns();
//     context->argnum = 0;

//     context->processor_id = (u16) bpf_get_smp_processor_id();

//     return 0;
// }



// static __always_inline int init_event_data(event_data_t *data, void *ctx) {
//     data->task = (struct task_struct *) bpf_get_current_task();
//     init_context(&data->context, data->task);
//     data->ctx = ctx;
//     return 0;
// }

// TODO: Needs to be optimized
struct event
{
    char comm[TASK_COMM_LEN];
    pid_t pid;
    pid_t tgid;
    pid_t ppid;
    uid_t uid;
    int retval;
    int args_count;
    unsigned int args_size;
    char args[FULL_MAX_ARGS_ARR];
};



static const struct event empty_event = {};

const volatile uid_t targ_uid = INVALID_UID;


// Definition of BPF Maps
BPF_HASH(execs, pid_t, struct event, 10240)
BPF_PERF_OUTPUT(events, 10240)
BPF_HASH(bridge, __u32, __u32, 10240)
BPF_RINGBUF(perfs, 1024)

static __always_inline bool valid_uid(uid_t uid)
{
    return uid != INVALID_UID;
}

SEC("tracepoint/syscalls/sys_enter_execve")
int tracepoint__syscalls__sys_enter_execve(struct trace_event_raw_sys_enter
                                               *ctx)
{
    u64 id;
    pid_t pid, tgid;
    unsigned int ret;

    struct event *event;
    struct task_struct *task;
    const char **args = (const char **)(ctx->args[1]);

    const char *argp;

    uid_t uid = (u32)bpf_get_current_uid_gid();

    id = bpf_get_current_pid_tgid();
    pid = (pid_t)id;
    tgid = id >> 32;

    if (bpf_map_update_elem(&execs, &pid, &empty_event, BPF_NOEXIST))
        return 0;

    event = bpf_map_lookup_elem(&execs, &pid);
    if (!event)
        return 0;

    event->pid = pid;
    event->tgid = tgid;
    event->uid = uid;
    task = (struct task_struct *)bpf_get_current_task();
    event->ppid = (pid_t)BPF_CORE_READ(task, real_parent, tgid);
    event->args_count = 0;
    event->args_size = 0;

    ret =
        bpf_probe_read_user_str(event->args, ARGSIZE,
                                (const char *)ctx->args[0]);
    if (ret <= ARGSIZE)
    {
        event->args[ret] = ' ';
        event->args_size += ret + 1;
    }
    else
    {
        /* write an empty string */
        event->args[0] = '\0';
        event->args_size++;
    }

    event->args_count++;
#pragma clang loop unroll(full)
    for (int i = 1; i < TOTAL_MAX_ARGS; i++)
    {
        bpf_probe_read_user(&argp, sizeof(argp), &args[i]);

        if (!argp)
            return 0;

        if (event->args_size > LAST_ARG)
            return 0;

        ret =
            bpf_probe_read_user_str(&event->args[event->args_size],
                                    ARGSIZE, argp);
        if (ret > ARGSIZE)
            return 0;

        event->args[ret] = ' ';
        event->args_count++;
        event->args_size += ret+1;
    }

    /* try to read one more argument to check if there is one */
    bpf_probe_read_user(&argp, sizeof(argp), &args[TOTAL_MAX_ARGS]);
    if (!argp)
        return 0;

    /* pointer to max_args+1 isn't null, assume we have more arguments */
    event->args_count++;

    return 0;
}

SEC("tracepoint/syscalls/sys_exit_execve")
int tracepoint__syscalls__sys_exit_execve(struct trace_event_raw_sys_exit *ctx)
{
    u64 id;
    pid_t pid;
    int ret;
    struct event *event;

    u32 uid = (u32)bpf_get_current_uid_gid();
    if (valid_uid(targ_uid) && targ_uid != uid)
        return 0;

    id = bpf_get_current_pid_tgid();
    pid = (pid_t)id;
    event = bpf_map_lookup_elem(&execs, &pid);
    if (!event)
        return 0;

    ret = ctx->ret;
    event->retval = ret;
    bpf_get_current_comm(&event->comm, sizeof(event->comm));
    size_t len = EVENT_SIZE(event);
    if (len <= sizeof(*event))
        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, event,
                              len);
    return 0;
}

SEC("xdp")
int xdp_proxy(struct xdp_md *ctx)
{
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    __u64 nh_off = 0;
    __u64 flags = BPF_F_CURRENT_CPU;

    struct ethhdr *eth = data;
    nh_off = sizeof(struct ethhdr);
    /* abort on illegal packets */
    if (data + nh_off > data_end)
    {
        return XDP_ABORTED;
    }

    /* do nothing for non-IP packets */
    if (eth->h_proto != bpf_htons(ETH_P_IP))
    {
        return XDP_PASS;
    }

    struct iphdr *iph = data + nh_off;
    nh_off += sizeof(struct iphdr);
    /* abort on illegal packets */
    if (data + nh_off > data_end)
    {
        return XDP_ABORTED;
    }

    /* do nothing for non-TCP packets */
    if (iph->protocol != IPPROTO_TCP)
    {
        return XDP_PASS;
    }

    struct tcphdr *tcphdr = data + nh_off;
    nh_off += sizeof(struct tcphdr);
    if (data + nh_off > data_end)
    {
        return XDP_ABORTED;
    }

    __u32 key = bpf_ntohl(iph->saddr);
    __u32 value = 0;
    char *payload = data + nh_off;

    if (bpf_map_lookup_elem(&bridge, &key) == NULL)
    {
        return XDP_PASS;
    }

    return XDP_DROP;
}

SEC("xdp")
int __test_trace_xdp(struct xdp_md *ctx) {
    int *lucknum;
    lucknum = bpf_ringbuf_reserve(&perfs, sizeof(int), 0);
    if (!lucknum)
    {
        return XDP_PASS;
    }

    *lucknum = 7070;
    bpf_ringbuf_submit(lucknum, 0);
    return XDP_PASS;
}

char LICENSE[] SEC("license") = "GPL";