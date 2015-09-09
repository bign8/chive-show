package pipeline

type SetmentConfig struct {
  InputStreams []InputStreamConfig
  OutputStreams []OutputStreamConfig
}

type InputStreamConfig struct {
  StreamName string
  KeyExtractor string
}

type OutputStreamConfig struct {
  StreamName string
  RecordFormat string
}
