package scanner

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

func Worker(ctx context.Context, urls <-chan string, results chan<- ResponseData, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case url, ok := <-urls:
			if !ok {
				return
			}

			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				fmt.Printf("Ошибка создания запроса для URL %s: %v\n", url, err)
				continue
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Printf("Ошибка выполнения запроса для URL %s: %v\n", url, err)
				continue
			}
			defer resp.Body.Close()

			select {
			case results <- ResponseData{
				Url: url,
				Headers: map[string]string{
					"Content-Type": resp.Header.Get("Content-Type"),
					"Server":       resp.Header.Get("Server"),
				},
				Status: resp.Status,
			}:
			case <-ctx.Done():
				fmt.Println("Контекст завершён, выходим из Worker")
				return
			}

		case <-ctx.Done():
			fmt.Println("Контекст завершён, выходим из Worker")
			return
		}
	}
}
