package services

type Service interface {
	BeforeService()
	DoService()
	AfterService()
}
