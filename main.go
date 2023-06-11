package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

const (
	projectId  = "inbound-planet-381014"      // GCPのプロジェクトID
	topicId    = "test_pubsub"                // pubsubのトピックID
	pubsubKey  = "keys/PubsubPublishKey.json" // GCPのjsonキー保存先
	bufferSize = 1000
	n          = 10 //メッセージ送信数
)

func publish_topic(
	ctx context.Context, //バックグランドで実行するためのcontext
	pubsubTopic *pubsub.Topic, //pubsubトピックのポインタ
	msg string, // pubsubへ送信するメッセージ内容
) {
	// 文字列をbytes変換する
	msgB := []byte(msg)
	// Pubsubにパブリッシュする
	msgP := &pubsub.Message{Data: msgB}
	_, err := pubsubTopic.Publish(ctx, msgP).Get(ctx)
	if err != nil {
		fmt.Println("Failed to publish response message: ", err)
		//他のエラーハンドリング
		return
	}
	fmt.Printf("パブリッシュされました%s", msg)
}

func main() {
	// pubsubでクライアント初期化
	ctx := context.Background()
	pubsub_client, err := pubsub.NewClient(ctx, projectId, option.WithCredentialsFile(pubsubKey))
	if err != nil {
		fmt.Printf("Failed to make pubsub client:%s ", err)
		return
	}
	defer pubsub_client.Close()

	pubsubTopic := pubsub_client.Topic(topicId)
	//10Kbytes毎に送信するバッチ設定
	//この処理を行っておくと高速にメッセージ送信でも10Kbytes毎にバルク送信してくれます
	pubsubTopic.PublishSettings.ByteThreshold = 10000

	var wg sync.WaitGroup
	// メッセージを作って送信する部分を並行処理で記述
	func() {
		for i := 0; i < n+1; i++ {
			wg.Add(1)
			msg := fmt.Sprintf(`
			"time": %s,
			"message": "Hello World %v"
            `, time.Now().Format("2006-01-02 15:04:05.123456"), i)

			go func() {
				defer wg.Done()
				publish_topic(ctx, pubsubTopic, msg)
			}()
		}
	}()
	wg.Wait()
}
