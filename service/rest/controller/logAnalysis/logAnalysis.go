package loganalysis

type LogAnalysisController struct {
	BaseUrl string
}

func NewLogAnalysisController() *LogAnalysisController {
	return &LogAnalysisController{
		BaseUrl: "http://192.168.176.1:7070",
	}
}
