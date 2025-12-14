# Research Outputs

## 1. operations.json
```json
{
  "IORING_OP_NOP": {
    "value": 0,
    "min_kernel": "5.1",
    "sqe_fields": ["user_data", "flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "medium"
  },
  "IORING_OP_READV": {
    "value": 1,
    "min_kernel": "5.1",
    "sqe_fields": ["fd", "addr", "len", "off", "user_data", "flags", "rw_flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "critical"
  },
  "IORING_OP_WRITEV": {
    "value": 2,
    "min_kernel": "5.1",
    "sqe_fields": ["fd", "addr", "len", "off", "user_data", "flags", "rw_flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "critical"
  },
  "IORING_OP_FSYNC": {
    "value": 3,
    "min_kernel": "5.1",
    "sqe_fields": ["fd", "user_data", "flags", "fsync_flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "high"
  },
  "IORING_OP_READ_FIXED": {
    "value": 4,
    "min_kernel": "5.1",
    "sqe_fields": ["fd", "addr", "len", "off", "user_data", "flags", "rw_flags", "buf_index"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_WRITE_FIXED": {
    "value": 5,
    "min_kernel": "5.1",
    "sqe_fields": ["fd", "addr", "len", "off", "user_data", "flags", "rw_flags", "buf_index"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_POLL_ADD": {
    "value": 6,
    "min_kernel": "5.5",
    "sqe_fields": ["fd", "user_data", "flags", "poll_events", "poll32_events", "len"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_POLL_REMOVE": {
    "value": 7,
    "min_kernel": "5.5",
    "sqe_fields": ["addr", "user_data", "flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_SYNC_FILE_RANGE": {
    "value": 8,
    "min_kernel": "5.1",
    "sqe_fields": ["fd", "off", "len", "user_data", "flags", "sync_range_flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "medium"
  },
  "IORING_OP_SENDMSG": {
    "value": 9,
    "min_kernel": "5.3",
    "sqe_fields": ["fd", "addr", "len", "user_data", "flags", "msg_flags", "addr_len"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "critical"
  },
  "IORING_OP_RECVMSG": {
    "value": 10,
    "min_kernel": "5.3",
    "sqe_fields": ["fd", "addr", "len", "user_data", "flags", "msg_flags", "addr_len", "buf_group"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "critical"
  },
  "IORING_OP_TIMEOUT": {
    "value": 11,
    "min_kernel": "5.4",
    "sqe_fields": ["addr", "user_data", "flags", "timeout_flags", "off"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "high"
  },
  "IORING_OP_TIMEOUT_REMOVE": {
    "value": 12,
    "min_kernel": "5.5",
    "sqe_fields": ["addr", "user_data", "flags", "timeout_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "high"
  },
  "IORING_OP_ACCEPT": {
    "value": 13,
    "min_kernel": "5.5",
    "sqe_fields": ["fd", "addr", "addr_len", "user_data", "flags", "accept_flags", "file_index"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "critical"
  },
  "IORING_OP_ASYNC_CANCEL": {
    "value": 14,
    "min_kernel": "5.5",
    "sqe_fields": ["addr", "user_data", "flags", "cancel_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_LINK_TIMEOUT": {
    "value": 15,
    "min_kernel": "5.5",
    "sqe_fields": ["addr", "user_data", "flags", "timeout_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_CONNECT": {
    "value": 16,
    "min_kernel": "5.5",
    "sqe_fields": ["fd", "addr", "addr_len", "user_data", "flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "critical"
  },
  "IORING_OP_FALLOCATE": {
    "value": 17,
    "min_kernel": "5.6",
    "sqe_fields": ["fd", "off", "addr", "len", "user_data", "flags", "fallocate_flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "medium"
  },
  "IORING_OP_OPENAT": {
    "value": 18,
    "min_kernel": "5.6",
    "sqe_fields": ["fd", "addr", "user_data", "flags", "open_flags", "mode", "addr2"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_CLOSE": {
    "value": 19,
    "min_kernel": "5.6",
    "sqe_fields": ["fd", "user_data", "flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "high"
  },
  "IORING_OP_FILES_UPDATE": {
    "value": 20,
    "min_kernel": "5.6",
    "sqe_fields": ["addr", "len", "off", "user_data", "flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "medium"
  },
  "IORING_OP_STATX": {
    "value": 21,
    "min_kernel": "5.6",
    "sqe_fields": ["fd", "addr", "addr2", "user_data", "flags", "statx_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "medium"
  },
  "IORING_OP_READ": {
    "value": 22,
    "min_kernel": "5.6",
    "sqe_fields": ["fd", "addr", "len", "off", "user_data", "flags", "rw_flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "critical"
  },
  "IORING_OP_WRITE": {
    "value": 23,
    "min_kernel": "5.6",
    "sqe_fields": ["fd", "addr", "len", "off", "user_data", "flags", "rw_flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "critical"
  },
  "IORING_OP_FADVISE": {
    "value": 24,
    "min_kernel": "5.6",
    "sqe_fields": ["fd", "off", "len", "user_data", "flags", "fadvise_advice"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "medium"
  },
  "IORING_OP_MADVISE": {
    "value": 25,
    "min_kernel": "5.6",
    "sqe_fields": ["addr", "len", "user_data", "flags", "fadvise_advice"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "medium"
  },
  "IORING_OP_SEND": {
    "value": 26,
    "min_kernel": "5.6",
    "sqe_fields": ["fd", "addr", "len", "user_data", "flags", "msg_flags", "addr_len"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "critical"
  },
  "IORING_OP_RECV": {
    "value": 27,
    "min_kernel": "5.6",
    "sqe_fields": ["fd", "addr", "len", "user_data", "flags", "msg_flags", "buf_group"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "critical"
  },
  "IORING_OP_OPENAT2": {
    "value": 28,
    "min_kernel": "5.6",
    "sqe_fields": ["fd", "addr", "len", "user_data", "flags", "open_flags", "addr2"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_EPOLL_CTL": {
    "value": 29,
    "min_kernel": "5.6",
    "sqe_fields": ["fd", "addr", "len", "user_data", "flags", "epoll_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "medium"
  },
  "IORING_OP_SPLICE": {
    "value": 30,
    "min_kernel": "5.7",
    "sqe_fields": ["fd", "off", "splice_fd_in", "len", "user_data", "flags", "splice_flags", "splice_off_in"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "medium"
  },
  "IORING_OP_PROVIDE_BUFFERS": {
    "value": 31,
    "min_kernel": "5.7",
    "sqe_fields": ["addr", "len", "buf_group", "user_data", "flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_REMOVE_BUFFERS": {
    "value": 32,
    "min_kernel": "5.7",
    "sqe_fields": ["len", "buf_group", "user_data", "flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_TEE": {
    "value": 33,
    "min_kernel": "5.8",
    "sqe_fields": ["fd", "off", "splice_fd_in", "len", "user_data", "flags", "splice_flags", "splice_off_in"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "medium"
  },
  "IORING_OP_SHUTDOWN": {
    "value": 34,
    "min_kernel": "5.11",
    "sqe_fields": ["fd", "len", "user_data", "flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "medium"
  },
  "IORING_OP_RENAMEAT": {
    "value": 35,
    "min_kernel": "5.11",
    "sqe_fields": ["fd", "addr", "addr2", "user_data", "flags", "rename_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "medium"
  },
  "IORING_OP_UNLINKAT": {
    "value": 36,
    "min_kernel": "5.11",
    "sqe_fields": ["fd", "addr", "user_data", "flags", "unlink_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "medium"
  },
  "IORING_OP_MKDIRAT": {
    "value": 37,
    "min_kernel": "5.15",
    "sqe_fields": ["fd", "addr", "user_data", "flags", "mode"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "low"
  },
  "IORING_OP_SYMLINKAT": {
    "value": 38,
    "min_kernel": "5.15",
    "sqe_fields": ["fd", "addr", "addr2", "user_data", "flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "low"
  },
  "IORING_OP_LINKAT": {
    "value": 39,
    "min_kernel": "5.15",
    "sqe_fields": ["fd", "addr", "addr2", "user_data", "flags", "hardlink_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "low"
  },
  "IORING_OP_MSG_RING": {
    "value": 40,
    "min_kernel": "5.10",
    "sqe_fields": ["fd", "len", "user_data", "flags", "msg_ring_flags", "addr", "off"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "high",
    "priority": "medium"
  },
  "IORING_OP_FSETXATTR": {
    "value": 41,
    "min_kernel": "5.15",
    "sqe_fields": ["fd", "addr", "addr2", "len", "user_data", "flags", "xattr_flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "low"
  },
  "IORING_OP_SETXATTR": {
    "value": 42,
    "min_kernel": "5.15",
    "sqe_fields": ["addr", "addr2", "len", "user_data", "flags", "xattr_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "low"
  },
  "IORING_OP_FGETXATTR": {
    "value": 43,
    "min_kernel": "5.15",
    "sqe_fields": ["fd", "addr", "addr2", "len", "user_data", "flags", "xattr_flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "low"
  },
  "IORING_OP_GETXATTR": {
    "value": 44,
    "min_kernel": "5.15",
    "sqe_fields": ["addr", "addr2", "len", "user_data", "flags", "xattr_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "low"
  },
  "IORING_OP_SOCKET": {
    "value": 45,
    "min_kernel": "5.19",
    "sqe_fields": ["len", "user_data", "flags", "domain", "type", "protocol", "rw_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_URING_CMD": {
    "value": 46,
    "min_kernel": "5.15",
    "sqe_fields": ["fd", "cmd_op", "user_data", "flags", "uring_cmd_flags", "cmd"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "high",
    "priority": "medium"
  },
  "IORING_OP_SEND_ZC": {
    "value": 47,
    "min_kernel": "5.18",
    "sqe_fields": ["fd", "addr", "len", "user_data", "flags", "msg_flags", "addr_len", "addr3"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "high",
    "priority": "high"
  },
  "IORING_OP_SENDMSG_ZC": {
    "value": 48,
    "min_kernel": "6.0",
    "sqe_fields": ["fd", "addr", "len", "user_data", "flags", "msg_flags", "addr_len"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "high",
    "priority": "high"
  },
  "IORING_OP_READ_MULTISHOT": {
    "value": 49,
    "min_kernel": "6.0",
    "sqe_fields": ["fd", "user_data", "flags", "rw_flags", "buf_group"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "high",
    "priority": "high"
  },
  "IORING_OP_WAITID": {
    "value": 50,
    "min_kernel": "6.0",
    "sqe_fields": ["addr", "user_data", "flags", "waitid_flags", "addr2"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "low"
  },
  "IORING_OP_FUTEX_WAIT": {
    "value": 51,
    "min_kernel": "6.4",
    "sqe_fields": ["addr", "user_data", "flags", "futex_flags", "off", "addr2", "addr3"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "high",
    "priority": "medium"
  },
  "IORING_OP_FUTEX_WAKE": {
    "value": 52,
    "min_kernel": "6.4",
    "sqe_fields": ["addr", "len", "user_data", "flags", "futex_flags", "off"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "high",
    "priority": "medium"
  },
  "IORING_OP_FUTEX_WAITV": {
    "value": 53,
    "min_kernel": "6.4",
    "sqe_fields": ["addr", "len", "user_data", "flags", "futex_flags", "off"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "high",
    "priority": "medium"
  },
  "IORING_OP_FIXED_FD_INSTALL": {
    "value": 54,
    "min_kernel": "6.3",
    "sqe_fields": ["fd", "user_data", "flags", "install_fd_flags", "file_index"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "medium"
  },
  "IORING_OP_FTRUNCATE": {
    "value": 55,
    "min_kernel": "6.11",
    "sqe_fields": ["fd", "off", "user_data", "flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "low",
    "priority": "medium"
  },
  "IORING_OP_BIND": {
    "value": 56,
    "min_kernel": "6.11",
    "sqe_fields": ["fd", "addr", "addr_len", "user_data", "flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_LISTEN": {
    "value": 57,
    "min_kernel": "6.11",
    "sqe_fields": ["fd", "len", "user_data", "flags"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_RECV_ZC": {
    "value": 58,
    "min_kernel": "6.12",
    "sqe_fields": ["fd", "addr", "len", "user_data", "flags", "msg_flags", "buf_group"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_BUFFER_SELECT", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "high",
    "priority": "high"
  },
  "IORING_OP_EPOLL_WAIT": {
    "value": 59,
    "min_kernel": "6.12",
    "sqe_fields": ["fd", "addr", "len", "user_data", "flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "high",
    "priority": "medium"
  },
  "IORING_OP_READV_FIXED": {
    "value": 60,
    "min_kernel": "6.12",
    "sqe_fields": ["fd", "addr", "len", "off", "user_data", "flags", "rw_flags", "buf_index"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_WRITEV_FIXED": {
    "value": 61,
    "min_kernel": "6.12",
    "sqe_fields": ["fd", "addr", "len", "off", "user_data", "flags", "rw_flags", "buf_index"],
    "flags_supported": ["IOSQE_FIXED_FILE", "IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "high"
  },
  "IORING_OP_PIPE": {
    "value": 62,
    "min_kernel": "6.12",
    "sqe_fields": ["fd", "len", "user_data", "flags", "pipe_flags"],
    "flags_supported": ["IOSQE_ASYNC", "IOSQE_IO_DRAIN", "IOSQE_IO_LINK", "IOSQE_IO_HARDLINK", "IOSQE_CQE_SKIP_SUCCESS"],
    "complexity": "medium",
    "priority": "medium"
  }
}
```

## 2. kernel_matrix.md

| Kernel Version | New Operations Added | New Flags Added | Structural Changes | Deprecated Features | Breaking Changes | Relevant IORING_FEAT_* Flags |
|----------------|----------------------|-----------------|---------------------|---------------------|------------------|------------------------------|
| 5.1 | NOP, READV, WRITEV, FSYNC, READ_FIXED, WRITE_FIXED | IORING_SETUP_IOPOLL, IORING_SETUP_SQPOLL, IORING_SETUP_SQ_AFF, IORING_SETUP_CQSIZE | Initial io_uring_sqe, cqe, params structs | None | None | IORING_FEAT_SINGLE_MMAP (introduced in 5.4) |
| 5.3 | SENDMSG, RECVMSG | IORING_SETUP_CLAMP, IORING_SETUP_ATTACH_WQ | Added addr_len in sqe unions | None | None | None |
| 5.4 | TIMEOUT | IORING_SETUP_R_DISABLED | None | None | None | IORING_FEAT_SINGLE_MMAP |
| 5.5 | POLL_ADD, POLL_REMOVE, ACCEPT, ASYNC_CANCEL, LINK_TIMEOUT, CONNECT | IORING_SETUP_SUBMIT_ALL | Added splice_fd_in | None | None | IORING_FEAT_NODROP |
| 5.6 | SYNC_FILE_RANGE, FALLOCATE, OPENAT, CLOSE, FILES_UPDATE, STATX, READ, WRITE, FADVISE, MADVISE, SEND, RECV, OPENAT2, EPOLL_CTL | IORING_SETUP_COOP_TASKRUN, IORING_SETUP_TASKRUN_FLAG | Added file_index, optlen | None | None | IORING_FEAT_CUR_PERSONALITY |
| 5.7 | SPLICE, PROVIDE_BUFFERS, REMOVE_BUFFERS | None | Added addr3 | None | None | IORING_FEAT_FAST_POLL |
| 5.8 | TEE | IORING_SETUP_SQE128, IORING_SETUP_CQE32 | None | None | None | None |
| 5.10 | MSG_RING | IORING_SETUP_SINGLE_ISSUER | Added xattr_flags | None | None | IORING_FEAT_NATIVE_WORKERS |
| 5.11 | SHUTDOWN, RENAMEAT, UNLINKAT | IORING_SETUP_DEFER_TASKRUN | None | None | None | None |
| 5.15 | MKDIRAT, SYMLINKAT, LINKAT, FSETXATTR, SETXATTR, FGETXATTR, GETXATTR, URING_CMD | IORING_SETUP_NO_MMAP | Added uring_cmd_flags | None | None | None |
| 5.18 | SEND_ZC | IORING_SETUP_REGISTERED_FD_ONLY | None | None | None | None |
| 5.19 | SOCKET | IORING_SETUP_NO_SQARRAY | Added waitid_flags | None | None | IORING_FEAT_REG_REG_RING |
| 6.0 | SENDMSG_ZC, READ_MULTISHOT, WAITID | None | Added futex_flags | None | None | None |
| 6.3 | FIXED_FD_INSTALL | None | Added install_fd_flags | None | None | None |
| 6.4 | FUTEX_WAIT, FUTEX_WAKE, FUTEX_WAITV | None | None | None | None | None |
| 6.8 | None (stable) | None | Added nop_flags in sqe union | None | None | None |
| 6.11 | FTRUNCATE, BIND, LISTEN | None | None | None | None | None |
| 6.12 | RECV_ZC, EPOLL_WAIT, READV_FIXED, WRITEV_FIXED, PIPE | IORING_SETUP_HYBRID_IOPOLL, IORING_SETUP_CQE_MIXED | Added zcrx_ifq_idx, write_stream, attr_ptr, pipe_flags | None | None | None |

## 3. api_design.md

### Type Definitions
```go
// Ring represents an io_uring instance
type Ring struct {
    fd int
    sq *SubmissionQueue
    cq *CompletionQueue
    params io_uring_params
    // ... mmap'd memory, etc.
}

// SubmissionQueue for SQEs
type SubmissionQueue struct {
    head, tail, mask, entries uint32
    flags uint32
    dropped uint32
    array []uint32
    sqes []io_uring_sqe
}

// CompletionQueue for CQEs
type CompletionQueue struct {
    head, tail, mask, entries uint32
    overflow uint32
    cqes []io_uring_cqe
}

// SQE mirrors struct io_uring_sqe
type SQE struct {
    Opcode uint8
    Flags uint8
    Ioprio uint16
    Fd int32
    Off uint64 // or Addr2
    Addr uint64 // or SpliceOffIn
    Len uint32
    // Union for op-specific
    OpFlags uint32 // rw_flags, etc.
    UserData uint64
    BufIndex uint16 // or BufGroup
    Personality uint16
    // Union: SpliceFdIn, FileIndex, etc.
    FileIndex int32
    // Union: Addr3, etc.
    Addr3 uint64
    // Cmd []byte for SQE128
}

// CQE mirrors struct io_uring_cqe
type CQE struct {
    UserData uint64
    Res int32
    Flags uint32
}
```

### Function Signatures
```go
// NewRing creates a new io_uring ring
func NewRing(entries uint32, params *io_uring_params) (*Ring, error)

// Submit submits SQEs and optionally waits for minComplete CQEs
func (r *Ring) Submit(minComplete uint32) (int, error)

// WaitCQE waits for at least one CQE
func (r *Ring) WaitCQE() (*CQE, error)

// PrepRead prepares an SQE for READ
func PrepRead(sqe *SQE, fd int, buf []byte, offset uint64)

// Similar Prep* for other ops

// RegisterBuffers registers fixed buffers
func (r *Ring) RegisterBuffers(buffers [][]byte) error

// RegisterFiles registers fixed files
func (r *Ring) RegisterFiles(fds []int) error
```

### Usage Examples
```go
ring, err := NewRing(128, &io_uring_params{Flags: IORING_SETUP_SQPOLL})
if err != nil { ... }

sqe := ring.GetSQE()
PrepRead(sqe, fd, buf, 0)
sqe.UserData = 123

submitted, err := ring.Submit(0)
cqe, err := ring.WaitCQE()
if cqe.Res < 0 { // error }
```

### Performance Considerations
- Use small ring sizes (e.g., 64-256) for low-latency workloads to minimize memory and cache misses.
- Enable SQPOLL for reduced syscall overhead but monitor CPU usage.
- Register buffers/files for zero-copy and lower overhead in repeated I/O.
- Batch submissions to amortize io_uring_enter calls.
- Avoid blocking waits; integrate with Go channels or runtime for non-blocking.
- For Go: Use unsafe.Pointer for mmap'd rings, but ensure no GC moves (pin memory).
- Monitor for CQ overflows; use larger CQ if needed.

## 4. implementation_order.md

### Phase 1: Core (ring setup, basic ops)
- Implement ring setup with io_uring_setup, mmap rings.
- Support basic flags: IORING_SETUP_IOPOLL, IORING_SETUP_SQPOLL, etc.
- Implement SQE/CQE structs and basic submission/completion.
- Basic ops: NOP, READV, WRITEV, FSYNC, READ, WRITE, TIMEOUT.
- Error handling and multi-kernel feature detection.

### Phase 2: Network (accept, connect, send, recv)
- Add network ops: ACCEPT, CONNECT, SEND, RECV, SENDMSG, RECVMSG.
- Support multishot for RECV, ACCEPT.
- Integrate zero-copy with SEND_ZC, RECV_ZC.
- Buffer selection and provision.
- SQPOLL integration for low-latency network.

### Phase 3: Advanced (SQPOLL, registered buffers)
- Full SQPOLL with kernel thread management.
- Registered buffers and files.
- Linked and drained operations.
- Multishot and fixed fd install.
- Performance optimizations: batching, no-mmap.

### Phase 4: Complete (all remaining operations)
- File system ops: OPENAT, CLOSE, STATX, FALLOCATE, RENAMEAT, etc.
- Advanced: EPOLL_CTL, FUTEX_*, MSG_RING, URING_CMD.
- XATTR, SOCKET, PIPE, BIND, LISTEN.
- Full flag support, including SQE128/CQE32.

## 5. test_plan.md

### Unit Test Structure
- Test individual Prep* functions for correct SQE field population.
- Mock kernel responses for CQE parsing and error handling.
- Test flag combinations (e.g., IOSQE_ASYNC, IOSQE_FIXED_FILE).
- Use Go's testing package with table-driven tests for ops.

### Integration Test Scenarios
- Basic read/write to files/sockets.
- Network server: accept connections, send/recv data.
- Multishot recv on sockets.
- Linked ops: write + fsync.
- Timeout and cancel scenarios.
- Registered buffers/files in loops.
- SQPOLL mode with high-load submissions.

### Benchmark Suite Design
- Compare with stdlib I/O (e.g., net/http vs io_uring server).
- Measure throughput (IOPS, req/sec), latency (p99), CPU usage.
- Vary ring sizes, batch sizes, workloads (file read, net echo).
- Use tools like fio, ab for loads.
- Profile allocations, syscalls with pprof.

### CI/CD Kernel Matrix
- Test on kernels 5.15, 6.1, 6.8, 6.11, latest.
- Use Docker/QEMU for multi-version envs.
- Matrix: kernel version x features (SQPOLL, multishot).
- Run unit/integration/bench on each.
- Fail if unsupported ops used on old kernels.

## Questions for Human Review
1. Should we support kernels older than 6.8? (increases complexity significantly)
2. Should we integrate with Go's netpoller or stay independent?
3. Should registered buffer management be automatic or manual?
4. How should we handle operations not supported on the running kernel?
5. Should reference implementations be in-repo or separate?
