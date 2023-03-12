package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/pflag"
)

var (
	content    = pflag.StringP("content", "c", "", "Origin Mirror format：{ \"hub-mirror\": [] }")
	maxContent = pflag.IntP("maxContent", "m", 10, "Maximum number of Original mirrors")
	username   = pflag.StringP("username", "u", "", "DockerHub Username")
	password   = pflag.StringP("password", "p", "", "DockerHub Password")
	outputPath = pflag.StringP("outputPath", "o", "output.sh", "DockerHub Mirror quickly pull shell script.")
)

var hubMirrors struct {
	Content []string `json:"hub-mirror"`
}

// check DockerHub username and password.
func checkUserPassword(cli *client.Client, username, password *string) (authStr string, err error) {
	if *username == "" || *password == "" {
		panic("username or password cannot be empty.")
	}
	authConfig := types.AuthConfig{
		Username: *username,
		Password: *password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	authStr = base64.URLEncoding.EncodeToString(encodedJSON)
	_, err = cli.RegistryLogin(context.Background(), authConfig)
	if err != nil {
		return "", err
	}
	return authStr, err
}

func log2File(output []struct {
	Source string
	Target string
}) {

	if len(output) == 0 {
		panic("output is empty.")
	}

	// 写入模板日志文件
	tmpl, err := template.New("pull_images").Parse(`{{- range . -}}

	docker pull {{ .Target }}
	docker tag {{ .Target }} {{ .Source }}

{{ end -}}`)
	if err != nil {
		panic(err)
	}
	outputFile, err := os.Create(*outputPath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	err = tmpl.Execute(outputFile, output)
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}

func main() {
	pflag.Parse()
	if pflag.NFlag() != 3 {
		fmt.Printf("missing parameters\n\n")
		pflag.PrintDefaults()
		return
	}

	fmt.Println("验证原始镜像内容")

	if err := json.Unmarshal([]byte(*content), &hubMirrors); err != nil {
		panic(err)
	}

	if len(hubMirrors.Content) > *maxContent {
		panic("content is too long.")
	}
	fmt.Printf("%+v\n", hubMirrors)

	fmt.Println("连接 Docker")
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	authStr, err := checkUserPassword(cli, username, password)
	if err != nil {
		panic(err)
	}

	fmt.Println("开始转换镜像")
	output := make([]struct {
		Source string
		Target string
	}, 0)

	wg := sync.WaitGroup{}

	for _, source := range hubMirrors.Content {
		if source == "" {
			continue
		}

		// 转换目标地址
		target := *username + "/" + strings.ReplaceAll(source, "/", ".")

		wg.Add(1)
		go func(source, target string) {
			defer wg.Done()

			fmt.Println("开始转换", source, "=>", target)
			ctx := context.Background()

			// 拉取镜像
			// 如果是私有仓库，需要在ImagePullOptions中传递RegistryAuth: authStr
			// pullOut, err := cli.ImagePull(ctx, source, types.ImagePullOptions{RegistryAuth: authStr})
			pullOut, err := cli.ImagePull(ctx, source, types.ImagePullOptions{})
			if err != nil {
				panic(err)
			}
			defer pullOut.Close()
			io.Copy(os.Stdout, pullOut)

			// 重新标签
			err = cli.ImageTag(ctx, source, target)
			if err != nil {
				panic(err)
			}

			// 上传镜像
			pushOut, err := cli.ImagePush(ctx, target, types.ImagePushOptions{
				RegistryAuth: authStr,
			})
			if err != nil {
				panic(err)
			}
			defer pushOut.Close()
			io.Copy(os.Stdout, pushOut)

			output = append(output, struct {
				Source string
				Target string
			}{Source: source, Target: target})
			fmt.Println("转换成功", source, "=>", target)
		}(source, target)
	}

	wg.Wait()

	log2File(output)
}
