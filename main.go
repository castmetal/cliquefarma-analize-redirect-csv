package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	inputhttp "github.com/castmetal/cliquefarma-analize-redirect-csv/http"
	"github.com/castmetal/cliquefarma-analize-redirect-csv/logger"
	"golang.org/x/time/rate"
)

type RowReader struct {
	chRow     chan []string
	csvwriter *csv.Writer
	mu        sync.Mutex
	ctx       context.Context
}

func NewRowReader(ctx context.Context, csvwriter *csv.Writer) RowReader {
	return RowReader{
		chRow:     make(chan []string, 200),
		csvwriter: csvwriter,
		mu:        sync.Mutex{},
		ctx:       ctx,
	}
}

func (r *RowReader) consumeRow() {
	for {
		select {
		case <-r.ctx.Done():
			return
		case row, ok := <-r.chRow:
			if !ok {
				return
			}

			var wg sync.WaitGroup

			if row[8] != "" && row[11] != "" {
				wg.Add(1)
				go func() {
					r.analyzeStatusAndWriteResponse(row[8], row[11], row)
					wg.Done()
				}()
			}

			if row[9] != "" && row[12] != "" {
				wg.Add(1)
				go func() {
					r.analyzeStatusAndWriteResponse(row[9], row[12], row)
					wg.Done()
				}()
			}

			if row[10] != "" && row[13] != "" {
				wg.Add(1)
				go func() {
					r.analyzeStatusAndWriteResponse(row[10], row[13], row)
					wg.Done()
				}()
			}

			wg.Wait()
		}
	}
}

func (r *RowReader) analyzeStatusAndWriteResponse(from string, to string, row []string) {
	var status string
	statusDe, statusPara := r.verifyUrls(from, to)

	if statusPara == 200 {
		status = "ANALISAR"
	} else if statusDe == 200 && statusPara != 200 {
		status = "REDIRECIONAR"
	} else {
		status = "ALTERAR"
	}

	strStatusDe := strconv.Itoa(statusDe)
	strStatusPara := strconv.Itoa(statusPara)

	r.mu.Lock()
	defer r.mu.Unlock()

	rowWritter := []string{
		fmt.Sprintf("%s\r", row[0]), fmt.Sprintf("%s\r", from), fmt.Sprintf("%s\r", to), fmt.Sprintf("%s\r", status), fmt.Sprintf("%s\r", strStatusDe), fmt.Sprintf("%s\r", strStatusPara),
	}
	_ = r.csvwriter.Write(rowWritter)

	r.csvwriter.Flush()
}

func (r *RowReader) verifyUrls(from string, to string) (int, int) {
	var status1 int
	var status2 int
	var wg sync.WaitGroup

	wg.Add(1)
	go func(requestStatus *int) {
		status, _ := FetchHttp(r.ctx, from)

		*requestStatus = status
		wg.Done()
	}(&status1)

	wg.Add(1)
	go func(requestStatus *int) {
		status, _ := FetchHttp(r.ctx, to)

		*requestStatus = status
		wg.Done()
	}(&status2)

	wg.Wait()

	return status1, status2
}

func main() {
	var csvFile *os.File
	var csvwriter *csv.Writer

	ctx := context.Background()

	file, err := os.Open("products_with_special_chars.csv")
	if err != nil {
		return
	}
	defer file.Close()

	csvFile, err = os.Create("output.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter = csv.NewWriter(csvFile)
	csvwriter.UseCRLF = true

	empRow := []string{
		"Sku", "De", "Para", "Status", "De Status", "Para Status",
	}
	_ = csvwriter.Write(empRow)

	csvwriter.Flush()
	defer csvFile.Close()

	reader := csv.NewReader(file)
	rowReader := NewRowReader(ctx, csvwriter)

	for i := 0; i <= 50; i++ {
		go rowReader.consumeRow()
	}

	limiter := rate.NewLimiter(rate.Limit(100), 1)
	i := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if i == 0 {
			i++
			continue
		}

		if err := limiter.Wait(ctx); err != nil {
			logger.Error(ctx, err, "rate limit exceeded")
			time.Sleep(5 * time.Second)
		}

		rowReader.chRow <- record
	}

	time.Sleep(90 * time.Second)
}

func FetchHttp(ctx context.Context, url string) (int, error) {
	var data io.ReadCloser
	status, err := GetStatusCodeHttp(ctx, url)
	if status >= 500 {
		data, status, _ = GetDataHttp(ctx, url, "GET")
	}

	if data != nil {
		data.Close()
	}

	return status, err
}

func GetStatusCodeHttp(ctx context.Context, url string) (int, error) {
	// Create an HTTP client
	client := http.Client{}

	// Create a HEAD request
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 500, err
	}

	// Send the request
	response, err := client.Do(request)
	if err != nil {
		return 500, err
	}

	return response.StatusCode, nil
}

func GetDataHttp(ctx context.Context, url string, method string) (io.ReadCloser, int, error) {
	if method == "" {
		method = "GET"
	}

	meta := map[string]interface{}{
		"targetURL": url,
		"method":    method,
	}

	client, err := inputhttp.New(ctx, meta)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, 500, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, 500, err
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusPartialContent:
		var buf bytes.Buffer
		length, err := io.Copy(&buf, res.Body)

		if err == nil && length <= 3 && length > 0 {
			res.Body.Close()
			buf.Reset()
			return nil, 404, errors.New("404 data, or not enough objects on this response")
		}

		closer := io.NopCloser(bytes.NewReader(buf.Bytes()))

		return closer, res.StatusCode, nil
	default:
		var buf bytes.Buffer
		_, err := io.Copy(&buf, res.Body)
		if err != nil {
			logger.Error(ctx, err, "could not read response body")
			res.Body.Close()
			return nil, res.StatusCode, fmt.Errorf("could not complete fetch: target: [%q] - response: [%q] - statusCode [%d]", url, buf.String(), res.StatusCode)
		}

		res.Body.Close()
		return nil, res.StatusCode, fmt.Errorf("could not complete fetch: target: [%q] - response: [%q] - statusCode [%d]", url, buf.String(), res.StatusCode)
	}
}
