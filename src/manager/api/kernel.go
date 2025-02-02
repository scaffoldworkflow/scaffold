package api

import (
	"errors"
	"fmt"
	"net/http"
	"scaffold/manager/kernel"
	"scaffold/manager/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	logger "github.com/jfcarter2358/go-logger"
)

type ExecuteMessage struct {
	Contents       string `json:"contents"`
	BlockIdxString string `json:"block_idx_str"`
	BlockIdx       int    `json:"block_idx"`
	KernelID       string `json:"kernel_id"`
	Language       string `json:"language"`
}

func ExecuteKernel(ctx *gin.Context) {
	var em ExecuteMessage
	if err := ctx.ShouldBindJSON(&em); err != nil {
		utils.Error(err, ctx, http.StatusInternalServerError)
		return
	}

	k := kernel.Kernels[em.KernelID]

	logger.Tracef("", "Kernel state: %v", k.Running)
	// if k.Running {
	// 	k.Kill()
	// }
	if k.Running {
		// logger.Tracef("", "Kernel is running, waiting...")
		// time.Sleep(1 * time.Second)
		utils.Error(errors.New("kernel is busy"), ctx, http.StatusUnprocessableEntity)
		return
	}

	k.PreviousRunID = k.RunID
	k.PreviousStdout = k.GetOutput()
	k.RunID = uuid.New().String()

	k.Channel <- fmt.Sprintf("marathon::lang_%s\n%s", em.Language, em.Contents)

	ctx.JSON(http.StatusOK, gin.H{"run_id": k.RunID})
}

type OutputMessage struct {
	Contents string `json:"contents"`
	Finished bool   `json:"finished"`
}

func GetKernelOutput(ctx *gin.Context) {
	kernelID := ctx.Param("kernel_id")
	runID := ctx.Param("run_id")

	k := kernel.Kernels[kernelID]

	switch runID {
	case k.RunID:
		ctx.JSON(http.StatusOK, OutputMessage{Contents: k.GetOutput(), Finished: !k.Running})
		return
	case k.PreviousRunID:
		ctx.JSON(http.StatusOK, OutputMessage{Contents: k.PreviousStdout, Finished: true})
		return
	}

	ctx.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("run does not exist with ID %s", runID)})
}
