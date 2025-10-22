package services

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/OrbitDeploy/OrbitDeploy/models"
	"gorm.io/gorm"
)

type QuadletData struct {
	Description      string
	ImageName        string
	ExecCommand      string
	AutoUpdatePolicy string
	PublishPorts     []string
	Volumes          []string
	EnvFilePath      string
}

func GenerateQuadletFileContent(db *gorm.DB, appName string, envFilePath string) (string, error) {
	var app models.Application

	err := db.Preload("ActiveRelease").Preload("Routings", "is_active = ?", true).Where("name = ?", appName).First(&app).Error
	if err != nil {
		return "", err
	}

	// 提取Volumes
	var volumes []string
	if app.Volumes.Data != nil {
		if volumesData, ok := app.Volumes.Data.([]interface{}); ok {
			for _, v := range volumesData {
				if vol, ok := v.(map[string]interface{}); ok {
					hostPath, hok := vol["host_path"].(string)
					containerPath, cok := vol["container_path"].(string)
					if hok && cok {
						volumes = append(volumes, hostPath+":"+containerPath)
					}
				}
			}
		}
	}

	// 提取PublishPorts
	var publishPorts []string
	for _, r := range app.Routings {
		publishPorts = append(publishPorts, fmt.Sprintf("%d:%d", r.HostPort, app.TargetPort))
	}

	// ExecCommand and AutoUpdatePolicy
	execCmd := ""
	if app.ExecCommand != nil {
		execCmd = *app.ExecCommand
	}
	autoUpdate := ""
	if app.AutoUpdatePolicy != nil {
		autoUpdate = *app.AutoUpdatePolicy
	}

	data := QuadletData{
		Description:      app.Description,
		ImageName:        app.ActiveRelease.ImageName,
		ExecCommand:      execCmd,
		AutoUpdatePolicy: autoUpdate,
		PublishPorts:     publishPorts,
		Volumes:          volumes,
		EnvFilePath:      envFilePath,
	}

	// 模板
	quadletTemplate := `[Unit]
Description={{ .Description }}
[Container]
Image={{ .ImageName }}
{{- if .ExecCommand }}
Exec={{ .ExecCommand }}
{{- end }}
{{- if .AutoUpdatePolicy }}
AutoUpdate={{ .AutoUpdatePolicy }}
{{- end }}
{{- range .PublishPorts }}
PublishPort={{ . }}
{{- end }}
{{- range .Volumes }}
Volume={{ . }}
{{- end }}
EnvironmentFile={{ .EnvFilePath }}
[Install]
WantedBy=default.target`

	tmpl, err := template.New("quadlet").Parse(quadletTemplate)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}
