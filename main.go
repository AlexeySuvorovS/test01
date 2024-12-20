package main


import (
   "bufio"
   "fmt"
   "os"
   "sync"
   "time"
)

type Logger interface {
   Log(message string) error
   Close() error
}

type FileLogger struct{ file *os.File }

func NewFileLogger(fileName string) (*FileLogger, error) {
   f, err := os.Create((fileName))
   if err != nil {
       return nil, err
   }

   return &FileLogger{file: f}, nil
}

func (f FileLogger) Log(message string) error {
   _, err := f.file.WriteString(message + "\n")
   return err
}

func (f FileLogger) Close() error {
   return f.file.Close()
}

type AsyncLogger struct {
   file    *os.File
   buf     *bufio.Writer
   mu      sync.Mutex
   closeCh chan bool
}

func NewAsyncFileLogger(fileName string) (*AsyncLogger, error) {
   f, err := os.Create(fileName)
   if err != nil {
       return nil, err
   }

   buf := bufio.NewWriter(f)

   ch := make(chan bool)

   al := AsyncLogger{
       file:    f,
       buf:     buf,
       closeCh: ch,
   }

   al.startWriter()

   return &al, nil
}

func (f *AsyncLogger) flushWriter() {
   f.mu.Lock()
   defer f.mu.Unlock()

   if f.buf.Size() > 0 {
       err := f.buf.Flush()

       if err == nil {
           f.buf.Reset(f.file)
       } else {
           fmt.Printf("ERROR: %s", err.Error())
       }
   }
}

func (f *AsyncLogger) startWriter() {
   go func(ch chan bool) {
       for {
           select {
           case <-ch:
               close(ch)
               return
           default:
               f.flushWriter()
           }
       }
   }(f.closeCh)
}

func (f *AsyncLogger) Log(message string) error {
   f.mu.Lock()
   defer f.mu.Unlock()

   _, err := f.buf.WriteString(message + "\n")

   return err
}

func (f *AsyncLogger) Close() error {
   defer f.file.Close()

   f.flushWriter()
   f.closeCh <- true

   return nil
}

func Worker(wg *sync.WaitGroup, name string, log Logger) {
   defer wg.Done()

   for i := 0; i < 5; i++ {
       log.Log(fmt.Sprintf("%s logging %d", name, i))
       time.Sleep(time.Second)
   }
}

func main() {
   fmt.Println("test 1")
   /*
       l, err := NewFileLogger("qwe001")
       if err != nil {
           fmt.Println(err.Error())
       }

       l.Log("str01")
       l.Log("str02")
       l.Log("str02")

       l.Close()
   */

   al, err := NewAsyncFileLogger("aaa01")
   if err != nil {
       fmt.Println(err.Error())
   }

   wg := &sync.WaitGroup{}

   wg.Add(1)
   go Worker(wg, "worker1", al)
   wg.Add(1)
   go Worker(wg, "worker2", al)
   wg.Wait()

   al.Close()
}
