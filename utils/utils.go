package utils

func AssembleCmdReq(cmd string) string {
	return cmd + "_REQ"
}

func AssembleCmdResp(cmd string) string {
	return cmd + "_RESP"
}
