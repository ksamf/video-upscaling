package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "D:\projects\GO\video-upscaling\video.proto"

	"google.golang.org/grpc"
)


func main() {
	// 1) Подключаемся к Python‑серверу
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("cannot dial server: %v", err)
	}
	defer conn.Close()
	client := pb.NewVideoProcessorClient(conn)

	// 2) Открываем видео-файл (или используем r.FormFile из HTTP)
	file, err := os.Open("some_video.mp4")
	if err != nil {
		log.Fatalf("cannot open video: %v", err)
	}
	defer file.Close()

	// 3) Создаём контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// 4) Стартуем клиентский стрим
	stream, err := client.ProcessVideo(ctx)
	if err != nil {
		log.Fatalf("cannot start stream: %v", err)
	}

	// 5) Читаем файл блоками и шлём их серверу
	buf := make([]byte, 1024*32) // 32KiB
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error reading file: %v", err)
		}
		chunk := &pb.VideoChunk{Data: buf[:n]}
		if err := stream.Send(chunk); err != nil {
			log.Fatalf("failed to send chunk: %v", err)
		}
	}

	// 6) Заканчиваем отправку и получаем ответ
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error on receive: %v", err)
	}
	fmt.Println("Server response:", res.GetMessage())
}
