package task

type collectorConf struct {
	Timeout int `json:",omitempty"`
	// file
	NoReopen bool   `json:",omitempty"`
	Mode     string `json:",omitempty"`
	FileName string `json:",omitempty"`
	// api
	URL string `json:",omitempty"`
	// api syslog
	Protocol string
	Addr     string `json:",omitempty"`

	End string
}

type parserConf struct {
	Mode       string
	Regex      string   `json:",omitempty"`
	Delimiters string   `json:",omitempty"`
	Columns    []string `json:",omitempty"`
}

type rewriterConf struct {
	Mode       string
	Column     string            `json:",omitempty"`
	Old        string            `json:",omitempty"`
	Value      string            `json:",omitempty"`
	Command    string            `json:",omitempty"`
	Delimiters string            `json:",omitempty"`
	Key        string            `json:",omitempty"`
	Columns    []string          `json:",omitempty"`
	Mapping    map[string]string `json:",omitempty"`
}

type validatorConf struct {
	Number int
	Mode   string
	Column string
	Type   string
	Value  string
	Regex  string
}

type handlerConf struct {
	// file
	Compress   bool `json:",omitempty"`
	MaxSize    int  `json:",omitempty"`
	MaxBackups int  `json:",omitempty"`
	MaxAge     int  `json:",omitempty"`
	// database
	Timeout int `json:",omitempty"`
	// common
	Mode     string
	Template string `json:",omitempty"`
	// file
	FileName string `json:",omitempty"`
	// database
	URI     string   `json:",omitempty"`
	Table   string   `json:",omitempty"`
	Columns []string `json:",omitempty"`
	Fields  []string `json:",omitempty"`

	Validators []validatorConf `json:",omitempty"`
}

// Conf Conf
type Conf struct {
	NoDegradation bool
	GoNum         int
	Collector     collectorConf
	Parser        parserConf
	Rewrites      []rewriterConf
	Validators    []validatorConf
	Handlers      []handlerConf
}
