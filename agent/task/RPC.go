package task

import (
	"context"
	pb "logagent/agent/task/plugins"
)

type server struct {
	pb.UnimplementedLoggerServer
	t *Task
}

// Module         int32                `protobuf:"varint,1,opt,name=module,proto3" json:"module,omitempty"`
// 	Timestamp      *timestamp.Timestamp `protobuf:"bytes,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
// 	TxnId          string               `protobuf:"bytes,3,opt,name=txn_id,json=txnId,proto3" json:"txn_id,omitempty"`
// 	AppId          string               `protobuf:"bytes,4,opt,name=app_id,json=appId,proto3" json:"app_id,omitempty"`
// 	DstIp          string               `protobuf:"bytes,5,opt,name=dst_ip,json=dstIp,proto3" json:"dst_ip,omitempty"`
// 	DstPort        int32                `protobuf:"varint,6,opt,name=dst_port,json=dstPort,proto3" json:"dst_port,omitempty"`
// 	SrcIp          string               `protobuf:"bytes,7,opt,name=src_ip,json=srcIp,proto3" json:"src_ip,omitempty"`
// 	SrcPort        int32                `protobuf:"varint,8,opt,name=src_port,json=srcPort,proto3" json:"src_port,omitempty"`
// 	Url            string               `protobuf:"bytes,9,opt,name=url,proto3" json:"url,omitempty"`
// 	Method         string               `protobuf:"bytes,10,opt,name=method,proto3" json:"method,omitempty"`
// 	Host           string               `protobuf:"bytes,11,opt,name=host,proto3" json:"host,omitempty"`
// 	HttpVersion    string               `protobuf:"bytes,12,opt,name=http_version,json=httpVersion,proto3" json:"http_version,omitempty"`
// 	ReqHeaders     string               `protobuf:"bytes,13,opt,name=req_headers,json=reqHeaders,proto3" json:"req_headers,omitempty"`
// 	ReqBody        []byte               `protobuf:"bytes,14,opt,name=req_body,json=reqBody,proto3" json:"req_body,omitempty"`
// 	ResCode        int32                `protobuf:"varint,15,opt,name=res_code,json=resCode,proto3" json:"res_code,omitempty"`
// 	ResHeaders     string               `protobuf:"bytes,16,opt,name=res_headers,json=resHeaders,proto3" json:"res_headers,omitempty"`
// 	ResBody        []byte               `protobuf:"bytes,17,opt,name=res_body,json=resBody,proto3" json:"res_body,omitempty"`
// 	RuleId         string               `protobuf:"bytes,18,opt,name=rule_id,json=ruleId,proto3" json:"rule_id,omitempty"`
// 	RuleVersion    string               `protobuf:"bytes,19,opt,name=rule_version,json=ruleVersion,proto3" json:"rule_version,omitempty"`
// 	Category       string               `protobuf:"bytes,20,opt,name=category,proto3" json:"category,omitempty"`
// TriggerMessage string

func (s *server) Emerg(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":        0,
		"module":          in.GetModule(),
		"timestamp":       in.GetTimestamp(),
		"txn_id":          in.GetTxnId(),
		"app_id":          in.GetAppId(),
		"dst_ip":          in.GetDstIp(),
		"dst_port":        in.GetDstPort(),
		"src_ip":          in.GetSrcIp(),
		"src_port":        in.GetSrcPort(),
		"url":             in.GetUrl(),
		"method":          in.GetMethod(),
		"host":            in.GetHost(),
		"http_version":    in.GetHttpVersion(),
		"req_headers":     in.GetReqHeaders(),
		"req_body":        in.GetReqBody(),
		"rule_id":         in.GetRuleId(),
		"rule_version":    in.GetRuleVersion(),
		"category":        in.GetCategory(),
		"trigger_message": in.GetTriggerMessage(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) Alert(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":        1,
		"module":          in.GetModule(),
		"timestamp":       in.GetTimestamp(),
		"txn_id":          in.GetTxnId(),
		"app_id":          in.GetAppId(),
		"dst_ip":          in.GetDstIp(),
		"dst_port":        in.GetDstPort(),
		"src_ip":          in.GetSrcIp(),
		"src_port":        in.GetSrcPort(),
		"url":             in.GetUrl(),
		"method":          in.GetMethod(),
		"host":            in.GetHost(),
		"http_version":    in.GetHttpVersion(),
		"req_headers":     in.GetReqHeaders(),
		"req_body":        in.GetReqBody(),
		"rule_id":         in.GetRuleId(),
		"rule_version":    in.GetRuleVersion(),
		"category":        in.GetCategory(),
		"trigger_message": in.GetTriggerMessage(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) Crit(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":        2,
		"module":          in.GetModule(),
		"timestamp":       in.GetTimestamp(),
		"txn_id":          in.GetTxnId(),
		"app_id":          in.GetAppId(),
		"dst_ip":          in.GetDstIp(),
		"dst_port":        in.GetDstPort(),
		"src_ip":          in.GetSrcIp(),
		"src_port":        in.GetSrcPort(),
		"url":             in.GetUrl(),
		"method":          in.GetMethod(),
		"host":            in.GetHost(),
		"http_version":    in.GetHttpVersion(),
		"req_headers":     in.GetReqHeaders(),
		"req_body":        in.GetReqBody(),
		"rule_id":         in.GetRuleId(),
		"rule_version":    in.GetRuleVersion(),
		"category":        in.GetCategory(),
		"trigger_message": in.GetTriggerMessage(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) Error(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":        3,
		"module":          in.GetModule(),
		"timestamp":       in.GetTimestamp(),
		"txn_id":          in.GetTxnId(),
		"app_id":          in.GetAppId(),
		"dst_ip":          in.GetDstIp(),
		"dst_port":        in.GetDstPort(),
		"src_ip":          in.GetSrcIp(),
		"src_port":        in.GetSrcPort(),
		"url":             in.GetUrl(),
		"method":          in.GetMethod(),
		"host":            in.GetHost(),
		"http_version":    in.GetHttpVersion(),
		"req_headers":     in.GetReqHeaders(),
		"req_body":        in.GetReqBody(),
		"rule_id":         in.GetRuleId(),
		"rule_version":    in.GetRuleVersion(),
		"category":        in.GetCategory(),
		"trigger_message": in.GetTriggerMessage(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) Warn(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":        4,
		"module":          in.GetModule(),
		"timestamp":       in.GetTimestamp(),
		"txn_id":          in.GetTxnId(),
		"app_id":          in.GetAppId(),
		"dst_ip":          in.GetDstIp(),
		"dst_port":        in.GetDstPort(),
		"src_ip":          in.GetSrcIp(),
		"src_port":        in.GetSrcPort(),
		"url":             in.GetUrl(),
		"method":          in.GetMethod(),
		"host":            in.GetHost(),
		"http_version":    in.GetHttpVersion(),
		"req_headers":     in.GetReqHeaders(),
		"req_body":        in.GetReqBody(),
		"rule_id":         in.GetRuleId(),
		"rule_version":    in.GetRuleVersion(),
		"category":        in.GetCategory(),
		"trigger_message": in.GetTriggerMessage(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) Notice(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":        5,
		"module":          in.GetModule(),
		"timestamp":       in.GetTimestamp(),
		"txn_id":          in.GetTxnId(),
		"app_id":          in.GetAppId(),
		"dst_ip":          in.GetDstIp(),
		"dst_port":        in.GetDstPort(),
		"src_ip":          in.GetSrcIp(),
		"src_port":        in.GetSrcPort(),
		"url":             in.GetUrl(),
		"method":          in.GetMethod(),
		"host":            in.GetHost(),
		"http_version":    in.GetHttpVersion(),
		"req_headers":     in.GetReqHeaders(),
		"req_body":        in.GetReqBody(),
		"rule_id":         in.GetRuleId(),
		"rule_version":    in.GetRuleVersion(),
		"category":        in.GetCategory(),
		"trigger_message": in.GetTriggerMessage(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) Info(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":        6,
		"module":          in.GetModule(),
		"timestamp":       in.GetTimestamp(),
		"txn_id":          in.GetTxnId(),
		"app_id":          in.GetAppId(),
		"dst_ip":          in.GetDstIp(),
		"dst_port":        in.GetDstPort(),
		"src_ip":          in.GetSrcIp(),
		"src_port":        in.GetSrcPort(),
		"url":             in.GetUrl(),
		"method":          in.GetMethod(),
		"host":            in.GetHost(),
		"http_version":    in.GetHttpVersion(),
		"req_headers":     in.GetReqHeaders(),
		"req_body":        in.GetReqBody(),
		"rule_id":         in.GetRuleId(),
		"rule_version":    in.GetRuleVersion(),
		"category":        in.GetCategory(),
		"trigger_message": in.GetTriggerMessage(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) Debug(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":        7,
		"module":          in.GetModule(),
		"timestamp":       in.GetTimestamp(),
		"txn_id":          in.GetTxnId(),
		"app_id":          in.GetAppId(),
		"dst_ip":          in.GetDstIp(),
		"dst_port":        in.GetDstPort(),
		"src_ip":          in.GetSrcIp(),
		"src_port":        in.GetSrcPort(),
		"url":             in.GetUrl(),
		"method":          in.GetMethod(),
		"host":            in.GetHost(),
		"http_version":    in.GetHttpVersion(),
		"req_headers":     in.GetReqHeaders(),
		"req_body":        in.GetReqBody(),
		"rule_id":         in.GetRuleId(),
		"rule_version":    in.GetRuleVersion(),
		"category":        in.GetCategory(),
		"trigger_message": in.GetTriggerMessage(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) EmergSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  0,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) AlertSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  1,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) CritSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  2,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) ErrorSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  3,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) WarnSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  4,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) NoticeSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  5,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) InfoSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  6,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) DebugSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  7,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

//  UserId    int64                `protobuf:"varint,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
// 	Ip        string               `protobuf:"bytes,4,opt,name=ip,proto3" json:"ip,omitempty"`
// 	Action    int64                `protobuf:"varint,5,opt,name=action,proto3" json:"action,omitempty"`
// 	Result    string               `protobuf:"bytes,6,opt,name=result,proto3" json:"result,omitempty"`
// 	Detail    string

func (s *server) EmergOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  0,
		"module":    9,
		"timestamp": in.GetTimestamp(),
		"user_id":   in.GetUserId(),
		"ip":        in.GetIp(),
		"action":    in.GetAction(),
		"result":    in.GetResult(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) AlertOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  1,
		"module":    9,
		"timestamp": in.GetTimestamp(),
		"user_id":   in.GetUserId(),
		"ip":        in.GetIp(),
		"action":    in.GetAction(),
		"result":    in.GetResult(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) CritOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  2,
		"module":    9,
		"timestamp": in.GetTimestamp(),
		"user_id":   in.GetUserId(),
		"ip":        in.GetIp(),
		"action":    in.GetAction(),
		"result":    in.GetResult(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) ErrorOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  3,
		"module":    9,
		"timestamp": in.GetTimestamp(),
		"user_id":   in.GetUserId(),
		"ip":        in.GetIp(),
		"action":    in.GetAction(),
		"result":    in.GetResult(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) WarnOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  4,
		"module":    9,
		"timestamp": in.GetTimestamp(),
		"user_id":   in.GetUserId(),
		"ip":        in.GetIp(),
		"action":    in.GetAction(),
		"result":    in.GetResult(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) NoticeOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  5,
		"module":    9,
		"timestamp": in.GetTimestamp(),
		"user_id":   in.GetUserId(),
		"ip":        in.GetIp(),
		"action":    in.GetAction(),
		"result":    in.GetResult(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) InfoOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  6,
		"module":    9,
		"timestamp": in.GetTimestamp(),
		"user_id":   in.GetUserId(),
		"ip":        in.GetIp(),
		"action":    in.GetAction(),
		"result":    in.GetResult(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) DebugOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  7,
		"module":    9,
		"timestamp": in.GetTimestamp(),
		"user_id":   in.GetUserId(),
		"ip":        in.GetIp(),
		"action":    in.GetAction(),
		"result":    in.GetResult(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}
