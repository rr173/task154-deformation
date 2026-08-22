package webui

import "strings"

func Document() string {
	sections := []string{
		"<!doctype html><html lang=zh-CN><head><meta charset=utf-8><meta name=viewport content='width=device-width,initial-scale=1'>",
		"<title>地基形变观测网平差台</title><style>" + Styles() + "</style></head><body>",
		"<main><header><div><p class=eyebrow>DEFORMATION NETWORK</p><h1>地基形变观测网平差台</h1></div><button id=demo>加载示例网络</button></header>",
		"<section class=summary><article><span>观测网络</span><strong id=networks>—</strong></article><article><span>待处理观测期</span><strong id=pending>—</strong></article><article><span>已发布成果</span><strong id=published>—</strong></article></section>",
		"<section class=layout><article class=panel><h2>网络概览</h2><p class=hint>网络和观测期从服务端 API 实时读取；结果发布后保持不可变。</p><div id=list class=list><p class=muted>正在读取数据…</p></div></article>",
		"<article class=panel><h2>操作说明</h2><ol><li>创建网络并导入固定点、待估点。</li><li>为观测期录入基线距离后执行平差。</li><li>撤回异常观测后重算，确认后发布成果版本。</li></ol><pre id=out aria-live=polite>等待操作</pre></article></section></main>",
		"<script>" + Script() + "</script></body></html>",
	}
	return strings.Join(sections, "")
}
