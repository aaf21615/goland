package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

// Параметры запроса
type query struct {
	url       string
	count     int
	errorText string
}

// Получить http и подсчитать подстроки "Go"
func processTask(q *query) {
	// отправляет запрос GET
	resp, err := http.Get(q.url)
	if err != nil {
		q.errorText = err.Error()
		return
	}

	// Оператор defer позволяет выполнить определенную функцию в конце программы,
	// при этом не важно, где в реальности вызывается эта функция.
	defer resp.Body.Close()

	// Читаем содержимое
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		q.errorText = err.Error()
		return
	}
	// Количество подстрок "Go"
	q.count = bytes.Count(body, []byte("Go"))
}

func main() {
	// Обрабатываются не более 5 url одновременно
	const maxWorkers = 5
	workers := 0

	// позволяет считывать данные с консоли
	scanner := bufio.NewScanner(os.Stdin)

	// Создаем и инициализируем каналы
	taskCh := make(chan query)
	doneCh := make(chan bool)
	processedDoneCh := make(chan bool)  // признак завершения
	processedTaskCh := make(chan query) // обработанные запросы
	var mutex sync.Mutex                // определяем мьютекс

	// анализировуем обработанные запросы и выводим в консоль
	go func() {
		totalCnt := 0 // Общее количестко
		mutex.Lock()  // блокируем доступ к переменной totalCnt
		for pTask := range processedTaskCh {
			errText := ""
			if pTask.errorText != "" {
				errText = "Ошибка: " + pTask.errorText
			}
			log.Printf("Количестко вхождений строки 'Go' в %s: %d %s \n", pTask.url, pTask.count, errText)
			totalCnt += pTask.count
		}
		log.Println("Сумма: ", totalCnt)
		mutex.Unlock()          // деблокируем доступ
		processedDoneCh <- true // Передаем данные в канал
	}()

	for scanner.Scan() {
		if workers < maxWorkers {
			// Создаем новый обработчик если нужно
			workers++
			go func() {
				// Получаем данные из канала и обрабатываем
				for query := range taskCh {
					processTask(&query)
					processedTaskCh <- query // Передаем данные в канал
				}
				doneCh <- true // Передаем данные в канал
			}()
		}
		url := scanner.Text()
		taskCh <- query{url: url} // Передаем данные в канал
	}

	if err := scanner.Err(); err != nil {
		log.Fatalln(err)
	}
	close(taskCh) // уведомить работников о том, что больше нет задач

	<-doneCh // получение данных из канала

	close(processedTaskCh)
	<-processedDoneCh // получение данных из канала
}
