package common

const (
	HEADER_SIZE       = 10
	MAGIC             = 0x06
	HANDLE_TYPE_SYN   = 0x01
	HANDLE_TYPE_CHAT  = 0x02
	HANDLE_TYPE_FINAL = 0x03
)

const (
	HANDLE_TYPE_EXPOSE_RES = 0x04
	HANDLE_TYPE_EXPOSE_REQ = 0x05
)

const (
	TRUE  = 0x01
	FALSE = 0x02
)
