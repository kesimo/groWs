package services

type InfoService struct {
}

func NewInfoService() *InfoService {
	return &InfoService{}
}

func (s *InfoService) GetServerInfo() map[string]interface{} {
	return map[string]interface{}{
		"version": "1.0.0",
		"server":  "grows",
		"slogan":  "A simple websocket framework for Go",
	}
}
