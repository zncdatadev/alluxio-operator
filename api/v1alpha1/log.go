package v1alpha1

type ContainerLoggingSpec struct {
	// +kubebuilder:validation:Optional
	Metastore *LoggingConfigSpec `json:"metastore,omitempty"`
}

type LoggingConfigSpec struct {
	// +kubebuilder:validation:Optional
	Loggers map[string]*LogLevelSpec `json:"loggers,omitempty"`

	// +kubebuilder:validation:Optional
	//if Console is set, the console appender will be added to the log4j.properties
	//file appender is the default appender, so we don't need to specify it
	Console *LogLevelSpec `json:"console,omitempty"`

	//// +kubebuilder:validation:Optional
	//default appender is file, so we don't need to specify it
	//File *LogLevelSpec `json:"file,omitempty"`
}

// LogLevelSpec
// level mapping example
// |---------------------|-----------------|
// |  superset log level |  zds log level  |
// |---------------------|-----------------|
// |  CRITICAL           |  FATAL          |
// |  ERROR              |  ERROR          |
// |  WARNING            |  WARN           |
// |  INFO               |  INFO           |
// |  DEBUG              |  DEBUG          |
// |  DEBUG              |  TRACE          |
// |---------------------|-----------------|
type LogLevelSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="INFO"
	// +kubebuilder:validation:Enum=FATAL;ERROR;WARN;INFO;DEBUG;TRACE
	Level string `json:"level,omitempty"`
}
