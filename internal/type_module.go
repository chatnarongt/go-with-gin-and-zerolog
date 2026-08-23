package internal

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type Module interface {
	Register(do.Injector, *gin.Engine) error
}

type OnModuleInit interface {
	OnModuleInit() error
}

type OnModuleDestroy interface {
	OnModuleDestroy(context.Context) error
}

type Controller interface {
	RegisterRoutes(*gin.Engine)
}

type Job func(context.Context, do.Injector) error

type JobRegistry interface {
	RegisterJob(name string, handler Job)
}

type JobRegistrar interface {
	RegisterJobs(JobRegistry)
}
