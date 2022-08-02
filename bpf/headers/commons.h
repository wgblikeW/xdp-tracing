#define BPF_MAP(_name, _type, _key_type, _value_type, _max_entries) \
    struct                                                          \
    {                                                               \
        __uint(type, _type);                                        \
        __uint(max_entries, _max_entries);                          \
        __type(key, _key_type);                                     \
        __type(value, _value_type);                                 \
    } _name SEC(".maps");

#define BPF_ARRAY(_name, _value_type, _max_entries) \
    BPF_MAP(_name, BPF_MAP_TYPE_ARRAY, u32, _value_type, _max_entries)

#define BPF_HASH(_name, _key_type, _value_type, _max_entries) \
    BPF_MAP(_name, BPF_MAP_TYPE_HASH, _key_type, _value_type, _max_entries)

#define BPF_PERF_OUTPUT(_name, _max_entries) \
    BPF_MAP(_name, BPF_MAP_TYPE_PERF_EVENT_ARRAY, u32, u32, _max_entries)

#define BPF_RINGBUF(_name, _max_entries)                  \
    struct                                                \
    {                                                     \
        __uint(type, BPF_MAP_TYPE_RINGBUF);               \
        __uint(max_entries, _max_entries);                \
    } _name SEC(".maps");