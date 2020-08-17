package task

import (
	"context"
	pb "logagent/agent/task/plugins"
)

type server struct {
	pb.UnimplementedLoggerServer
	t *Task
}

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

func (s *server) Crit(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
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

func (s *server) Warn(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
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

func (s *server) Notice(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
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

func (s *server) Info(ctx context.Context, in *pb.ProtectionLog) (*pb.Reply, error) {
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

func (s *server) CritSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  1,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) WarnSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  2,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) NoticeSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  3,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

func (s *server) InfoSysLog(ctx context.Context, in *pb.SystemLog) (*pb.Reply, error) {
	s.t.msgs <- map[string]interface{}{
		"severity":  4,
		"module":    8,
		"timestamp": in.GetTimestamp(),
		"event":     in.GetEvent(),
		"detail":    in.GetDetail(),
	}
	return &pb.Reply{Code: 0, Message: "successfully"}, nil
}

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

func (s *server) CritOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
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

func (s *server) WarnOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
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

func (s *server) NoticeOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
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

func (s *server) InfoOpLog(ctx context.Context, in *pb.OperationLog) (*pb.Reply, error) {
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
