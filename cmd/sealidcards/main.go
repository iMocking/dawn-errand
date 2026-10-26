// cmd/sealidcards 存量治理：把 riders 表中未加密（无 v1: 前缀）的身份证号批量加密回写。
// 用法：go run ./cmd/sealidcards            演练（只统计不写入）
//
//	go run ./cmd/sealidcards --apply    实际执行
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	"dawn-errand/internal/assembly"
	"dawn-errand/internal/config"
	"dawn-errand/internal/infrastructure/crypto"
)

func main() {
	apply := len(os.Args) > 1 && os.Args[1] == "--apply"
	ctx := context.Background()
	cfg := config.Load()
	app, err := assembly.Build(ctx, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "装配失败:", err)
		os.Exit(1)
	}
	defer app.Client.Close()
	cipher := crypto.New(cfg.ConfigCipherKey)

	rows, err := app.Client.Rider.Query().All(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查询失败:", err)
		os.Exit(1)
	}
	var plain, sealed int
	for _, e := range rows {
		if e.IDCard == "" || strings.HasPrefix(e.IDCard, "v1:") {
			continue
		}
		plain++
		if !apply {
			continue
		}
		value, err := cipher.Encrypt(e.IDCard)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加密失败 rider=%s: %v\n", e.ID, err)
			os.Exit(1)
		}
		if err := app.Client.Rider.UpdateOneID(e.ID).SetIDCard(value).Exec(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "回写失败 rider=%s: %v\n", e.ID, err)
			os.Exit(1)
		}
		sealed++
	}
	mode := "演练"
	if apply {
		mode = "已执行"
	}
	fmt.Printf("%s：发现明文身份证 %d 条，成功加密回写 %d 条\n", mode, plain, sealed)
}
