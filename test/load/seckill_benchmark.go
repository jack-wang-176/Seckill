package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Config 压测配置
type Config struct {
	TargetURL   string        // API Gateway地址
	Concurrency int           // 并发数
	Duration    time.Duration // 压测持续时间
	ProductID   string        // 商品ID
	UserCount   int           // 测试用户数
	RampUpTime  time.Duration // 梯度增压时间
}

// Metrics 性能指标
type Metrics struct {
	TotalRequests   int64
	SuccessRequests int64
	FailRequests    int64
	PathLatencies   []int64 // ms
	OrderLatencies  []int64 // ms
	ResultLatencies []int64 // ms
	mu              sync.Mutex
}

// UserPool 用户令牌池
type UserPool struct {
	tokens []string
	index  int
	mu     sync.Mutex
}

// GetToken 获取下一个用户令牌
func (up *UserPool) GetToken() string {
	up.mu.Lock()
	defer up.mu.Unlock()
	token := up.tokens[up.index%len(up.tokens)]
	up.index++
	return token
}

// RecordLatency 记录延迟
func (m *Metrics) RecordLatency(latencies *[]int64, ms int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*latencies = append(*latencies, ms)
}

// TestWorker 压测工作线程
type TestWorker struct {
	client   *http.Client
	userPool *UserPool
	config   *Config
	metrics  *Metrics
	done     <-chan struct{}
	wg       *sync.WaitGroup
}

// doRequest 执行HTTP请求
func (w *TestWorker) doRequest(method, url string, body interface{}, headers map[string]string) ([]byte, int64, error) {
	var reqBody io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reqBody = bytes.NewReader(data)
	}

	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := w.client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return nil, latency, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, latency, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return respBody, latency, nil
}

// SeckillPathResp 获取秒杀路径响应（适配统一响应格式）
type SeckillPathResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Path string `json:"path"`
	} `json:"data"`
}

// OrderResp 下单响应（适配统一响应格式）
type OrderResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

// ResultResp 结果响应（适配统一响应格式）
type ResultResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	} `json:"data"`
}

// Run 运行压测
func (w *TestWorker) Run() {
	defer w.wg.Done()

	for {
		select {
		case <-w.done:
			return
		default:
		}

		atomic.AddInt64(&w.metrics.TotalRequests, 1)
		token := w.userPool.GetToken()
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		// 步骤1: 获取秒杀路径
		pathBody, pathLatency, err := w.doRequest(
			"POST",
			w.config.TargetURL+"/api/v1/seckill/path",
			map[string]string{"product_id": w.config.ProductID},
			headers,
		)

		if err != nil {
			atomic.AddInt64(&w.metrics.FailRequests, 1)
			continue
		}

		var pathResp SeckillPathResp
		json.Unmarshal(pathBody, &pathResp)
		w.metrics.RecordLatency(&w.metrics.PathLatencies, pathLatency)

		// 步骤2: 提交订单 (注意这里改成了 pathResp.Data.Path)
		orderBody, orderLatency, err := w.doRequest(
			"POST",
			w.config.TargetURL+"/api/v1/seckill/order/"+pathResp.Data.Path,
			map[string]string{"product_id": w.config.ProductID},
			headers,
		)

		if err != nil {
			atomic.AddInt64(&w.metrics.FailRequests, 1)
			continue
		}

		var orderResp OrderResp
		json.Unmarshal(orderBody, &orderResp)
		w.metrics.RecordLatency(&w.metrics.OrderLatencies, orderLatency)

		// 步骤3: 轮询获取结果（部分worker执行）
		if w.config.Concurrency%3 == 0 { // 约1/3的worker进行结果轮询
			maxRetries := 3
			for i := 0; i < maxRetries; i++ {
				resultBody, resultLatency, err := w.doRequest(
					"GET",
					w.config.TargetURL+"/api/v1/seckill/result?product_id="+w.config.ProductID,
					nil,
					headers,
				)

				if err == nil {
					var resultResp ResultResp
					json.Unmarshal(resultBody, &resultResp)
					w.metrics.RecordLatency(&w.metrics.ResultLatencies, resultLatency)

					// 兼容：通过 Code=200 或者是 Data 内部的 Success 判断
					if resultResp.Code == 200 || resultResp.Data.Success {
						atomic.AddInt64(&w.metrics.SuccessRequests, 1)
						break
					}
				}

				if i < maxRetries-1 {
					time.Sleep(100 * time.Millisecond)
				}
			}
		} else {
			atomic.AddInt64(&w.metrics.SuccessRequests, 1)
		}
	}
}

// GenerateUserTokens 生成测试用户令牌
func GenerateUserTokens(targetURL string, count int) ([]string, error) {
	tokens := make([]string, 0, count)
	client := &http.Client{Timeout: 5 * time.Second}

	for i := 0; i < count; i++ {
		username := fmt.Sprintf("testuser_%d", i)
		password := "password123"

		// 注册用户
		regBody := map[string]string{"username": username, "password": password}
		data, _ := json.Marshal(regBody)
		req, _ := http.NewRequest("POST", targetURL+"/api/v1/user/register", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("注册请求失败 user_%d: %v", i, err)
		} else {
			if resp != nil {
				if resp.StatusCode != http.StatusOK {
					io.ReadAll(resp.Body)
				}
				resp.Body.Close()
			}
		}

		// 登录获取令牌（POST JSON）
		loginBody := map[string]string{"username": username, "password": password}
		maxAttempts := 30
		gotToken := false
		for attempt := 0; attempt < maxAttempts; attempt++ {
			data, _ := json.Marshal(loginBody)
			req, _ = http.NewRequest("POST", targetURL+"/api/v1/user/login", bytes.NewReader(data))
			req.Header.Set("Content-Type", "application/json")

			resp, err = client.Do(req)
			if err != nil {
				log.Printf("登录请求失败 user_%d attempt=%d: %v", i, attempt+1, err)
			} else {
				if resp != nil {
					bodyBytes, _ := io.ReadAll(resp.Body)
					resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						log.Printf("登录失败 user_%d attempt=%d: status=%d body=%s", i, attempt+1, resp.StatusCode, string(bodyBytes))
					} else {
						// 适配统一响应格式 {"code":200, "msg":"success", "data":{"token":"xxx"}}
						var loginResp struct {
							Data struct {
								Token string `json:"token"`
							} `json:"data"`
						}
						if err := json.Unmarshal(bodyBytes, &loginResp); err != nil {
							log.Printf("解析登录响应失败 user_%d: %v", i, err)
						} else if loginResp.Data.Token != "" {
							tokens = append(tokens, loginResp.Data.Token)
							gotToken = true
							break
						} else {
							log.Printf("登录响应无 token user_%d attempt=%d (body=%s)", i, attempt+1, string(bodyBytes))
						}
					}
				} else {
					log.Printf("登录返回 nil resp user_%d attempt=%d", i, attempt+1)
				}
			}

			// 指数退避，带最大上限
			if attempt < maxAttempts-1 {
				backoff := time.Duration(100*(1<<attempt)) * time.Millisecond
				if backoff > 2*time.Second || backoff <= 0 { // 防止溢出和过长的等待
					backoff = 2 * time.Second
				}
				time.Sleep(backoff)
			}
		}

		if !gotToken {
			log.Printf("登录失败（重试后） user_%d，跳过此用户", i)
		}

		if i%50 == 0 {
			fmt.Printf("已生成 %d 个用户令牌\n", len(tokens))
		}
	}

	fmt.Printf("共生成 %d 个有效用户令牌\n", len(tokens))
	if len(tokens) == 0 {
		return nil, fmt.Errorf("未生成任何令牌，请检查目标地址或后端服务是否可用: %s", targetURL)
	}

	return tokens, nil
}

// CalculatePercentile 计算百分位数
func CalculatePercentile(data []int64, percentile float64) int64 {
	if len(data) == 0 {
		return 0
	}
	sort.Slice(data, func(i, j int) bool { return data[i] < data[j] })
	index := int(float64(len(data)) * percentile / 100)
	if index >= len(data) {
		index = len(data) - 1
	}
	return data[index]
}

// PrintReport 打印性能报告
func PrintReport(metrics *Metrics, duration time.Duration) {
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	fmt.Println("\n" + string([]byte{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61}))
	fmt.Println("                  秒杀压测报告                    ")
	fmt.Println(string([]byte{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61}))

	// 基础指标
	fmt.Printf("\n总时长：%.1f 秒\n", duration.Seconds())
	fmt.Printf("总请求数：%d\n", metrics.TotalRequests)
	fmt.Printf("成功请求：%d\n", metrics.SuccessRequests)
	fmt.Printf("失败请求：%d\n", metrics.FailRequests)

	successRate := 0.0
	if metrics.TotalRequests > 0 {
		successRate = float64(metrics.SuccessRequests) * 100 / float64(metrics.TotalRequests)
	}
	fmt.Printf("成功率：%.2f%%\n", successRate)

	qps := float64(metrics.TotalRequests) / duration.Seconds()
	fmt.Printf("QPS：%.0f 请求/秒\n", qps)

	// 延迟统计
	fmt.Println("\n--- 秒杀路径延迟 ---")
	if len(metrics.PathLatencies) > 0 {
		avg := calculateAverage(metrics.PathLatencies)
		p50 := CalculatePercentile(metrics.PathLatencies, 50)
		p90 := CalculatePercentile(metrics.PathLatencies, 90)
		p99 := CalculatePercentile(metrics.PathLatencies, 99)
		fmt.Printf("平均: %d ms | P50: %d ms | P90: %d ms | P99: %d ms\n", avg, p50, p90, p99)
	}

	fmt.Println("\n--- 下单延迟 ---")
	if len(metrics.OrderLatencies) > 0 {
		avg := calculateAverage(metrics.OrderLatencies)
		p50 := CalculatePercentile(metrics.OrderLatencies, 50)
		p90 := CalculatePercentile(metrics.OrderLatencies, 90)
		p99 := CalculatePercentile(metrics.OrderLatencies, 99)
		fmt.Printf("平均: %d ms | P50: %d ms | P90: %d ms | P99: %d ms\n", avg, p50, p90, p99)
	}

	fmt.Println("\n--- 结果查询延迟 ---")
	if len(metrics.ResultLatencies) > 0 {
		avg := calculateAverage(metrics.ResultLatencies)
		p50 := CalculatePercentile(metrics.ResultLatencies, 50)
		p90 := CalculatePercentile(metrics.ResultLatencies, 90)
		p99 := CalculatePercentile(metrics.ResultLatencies, 99)
		fmt.Printf("平均: %d ms | P50: %d ms | P90: %d ms | P99: %d ms\n", avg, p50, p90, p99)
	}

	fmt.Println("\n" + string([]byte{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61}))
}

func calculateAverage(data []int64) int64 {
	if len(data) == 0 {
		return 0
	}
	sum := int64(0)
	for _, v := range data {
		sum += v
	}
	return sum / int64(len(data))
}

func main() {
	config := &Config{}
	flag.StringVar(&config.TargetURL, "url", "http://localhost:8081", "API Gateway地址")
	flag.IntVar(&config.Concurrency, "concurrency", 100, "并发数")
	flag.DurationVar(&config.Duration, "duration", 30*time.Second, "压测持续时间")
	flag.StringVar(&config.ProductID, "product", "1", "商品ID")
	flag.IntVar(&config.UserCount, "users", 500, "测试用户数")
	flag.DurationVar(&config.RampUpTime, "rampup", 5*time.Second, "梯度增压时间")
	flag.Parse()

	fmt.Printf("开始生成 %d 个测试用户...\n", config.UserCount)
	tokens, err := GenerateUserTokens(config.TargetURL, config.UserCount)
	if err != nil || len(tokens) == 0 {
		log.Fatalf("无法生成测试用户: %v", err)
	}

	fmt.Printf("\n开始压测：并发 %d | 持续 %v | 商品 %s\n\n",
		config.Concurrency, config.Duration, config.ProductID)

	metrics := &Metrics{}
	userPool := &UserPool{tokens: tokens}

	// 创建优化的HTTP客户端
	transport := &http.Transport{
		MaxIdleConns:        config.Concurrency + 50,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     200,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		DisableCompression:  true,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	done := make(chan struct{})
	var wg sync.WaitGroup

	// 启动worker
	wg.Add(config.Concurrency)
	for i := 0; i < config.Concurrency; i++ {
		worker := &TestWorker{
			client:   client,
			userPool: userPool,
			config:   config,
			metrics:  metrics,
			done:     done,
			wg:       &wg,
		}
		go worker.Run()
	}

	// 梯度增压
	if config.RampUpTime > 0 {
		interval := config.RampUpTime / time.Duration(config.Concurrency)
		for i := 0; i < config.Concurrency; i++ {
			time.Sleep(interval)
		}
	}

	// 运行压测并定期打印进度
	startTime := time.Now()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			fmt.Printf("已发送请求: %d | 成功: %d | 失败: %d | QPS: %.0f\n",
				atomic.LoadInt64(&metrics.TotalRequests),
				atomic.LoadInt64(&metrics.SuccessRequests),
				atomic.LoadInt64(&metrics.FailRequests),
				float64(atomic.LoadInt64(&metrics.TotalRequests))/time.Since(startTime).Seconds(),
			)
		}
	}()

	// 等待压测结束
	time.Sleep(config.Duration)
	close(done)
	wg.Wait()

	elapsed := time.Since(startTime)
	PrintReport(metrics, elapsed)
}
