package service

import "go.uber.org/zap"

func (s *Service) logRelay(txHash string) {
	s.l.Info("successfully relayed transaction", zap.String("tx.hash", txHash))
}
