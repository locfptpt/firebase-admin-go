// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package snippets

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

	"firebase.google.com/go"
	"firebase.google.com/go/links"
)

func testLinkStats() {
	// [START get_link_stats]
	client, err := app.DynamicLinks(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	stats, err := client.LinkStats(ctx, "https://abc.app.goo.gl/abc12",
		&StatOptions{lastNDays: 7})
	if err != nil {
		log.Fatalln(err)
	}
	for _, stat := range stats {
		if stat.Platform == links.Android && stat.EventType == links.Client {
			fmt.Printf("There were %v clicks on Android in the last 7 days", stat.Count)
		}
	}
	// [END get_link_stats]
}
func sendToToken(app *firebase.App) {
	// [START send_to_token_golang]
	// Obtain a messaging.Client from the App.
	ctx := context.Background()
	client, err := app.Messaging(ctx)

	// This registration token comes from the client FCM SDKs.
	registrationToken := "YOUR_REGISTRATION_TOKEN"

	// See documentation on defining a message payload.
	message := &messaging.Message{
		Data: map[string]string{
			"score": "850",
			"time":  "2:45",
		},
		Token: registrationToken,
	}

	// Send a message to the device corresponding to the provided
	// registration token.
	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
	// [END send_to_token_golang]
}

func sendToTopic(ctx context.Context, client *messaging.Client) {
	// [START send_to_topic_golang]
	// The topic name can be optionally prefixed with "/topics/".
	topic := "highScores"

	// See documentation on defining a message payload.
	message := &messaging.Message{
		Data: map[string]string{
			"score": "850",
			"time":  "2:45",
		},
		Topic: topic,
	}

	// Send a message to the devices subscribed to the provided topic.
	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
	// [END send_to_topic_golang]
}

func sendToCondition(ctx context.Context, client *messaging.Client) {
	// [START send_to_condition_golang]
	// Define a condition which will send to devices which are subscribed
	// to either the Google stock or the tech industry topics.
	condition := "'stock-GOOG' in topics || 'industry-tech' in topics"

	// See documentation on defining a message payload.
	message := &messaging.Message{
		Data: map[string]string{
			"score": "850",
			"time":  "2:45",
		},
		Condition: condition,
	}

	// Send a message to devices subscribed to the combination of topics
	// specified by the provided condition.
	response, err := client.Send(ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	// Response is a message ID string.
	fmt.Println("Successfully sent message:", response)
	// [END send_to_condition_golang]
}

func sendDryRun(ctx context.Context, client *messaging.Client) {
	message := &messaging.Message{
		Data: map[string]string{
			"score": "850",
			"time":  "2:45",
		},
		Token: "token",
	}

	// [START send_dry_run_golang]
	// Send a message in the dry run mode.
	response, err := client.SendDryRun(ctx, message)
	if err != nil {
		log.Fatalln(err)
	}
	// Response is a message ID string.
	fmt.Println("Dry run successful:", response)
	// [END send_dry_run_golang]
}

func androidMessage() *messaging.Message {
	// [START android_message_golang]
	oneHour := time.Duration(1) * time.Hour
	message := &messaging.Message{
		Android: &messaging.AndroidConfig{
			TTL:      &oneHour,
			Priority: "normal",
			Notification: &messaging.AndroidNotification{
				Title: "$GOOG up 1.43% on the day",
				Body:  "$GOOG gained 11.80 points to close at 835.67, up 1.43% on the day.",
				Icon:  "stock_ticker_update",
				Color: "#f45342",
			},
		},
		Topic: "industry-tech",
	}
	// [END android_message_golang]
	return message
}

func apnsMessage() *messaging.Message {
	// [START apns_message_golang]
	badge := 42
	message := &messaging.Message{
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: "$GOOG up 1.43% on the day",
						Body:  "$GOOG gained 11.80 points to close at 835.67, up 1.43% on the day.",
					},
					Badge: &badge,
				},
			},
		},
		Topic: "industry-tech",
	}
	// [END apns_message_golang]
	return message
}

func webpushMessage() *messaging.Message {
	// [START webpush_message_golang]
	message := &messaging.Message{
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: "$GOOG up 1.43% on the day",
				Body:  "$GOOG gained 11.80 points to close at 835.67, up 1.43% on the day.",
				Icon:  "https://my-server/icon.png",
			},
		},
		Topic: "industry-tech",
	}
	// [END webpush_message_golang]
	return message
}

func allPlatformsMessage() *messaging.Message {
	// [START multi_platforms_message_golang]
	oneHour := time.Duration(1) * time.Hour
	badge := 42
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: "$GOOG up 1.43% on the day",
			Body:  "$GOOG gained 11.80 points to close at 835.67, up 1.43% on the day.",
		},
		Android: &messaging.AndroidConfig{
			TTL: &oneHour,
			Notification: &messaging.AndroidNotification{
				Icon:  "stock_ticker_update",
				Color: "#f45342",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Badge: &badge,
				},
			},
		},
		Topic: "industry-tech",
	}
	// [END multi_platforms_message_golang]
	return message
}

func subscribeToTopic(ctx context.Context, client *messaging.Client) {
	topic := "highScores"

	// [START subscribe_golang]
	// These registration tokens come from the client FCM SDKs.
	registrationTokens := []string{
		"YOUR_REGISTRATION_TOKEN_1",
		// ...
		"YOUR_REGISTRATION_TOKEN_n",
	}

	// Subscribe the devices corresponding to the registration tokens to the
	// topic.
	response, err := client.SubscribeToTopic(ctx, registrationTokens, topic)
	if err != nil {
		log.Fatalln(err)
	}
	// See the TopicManagementResponse reference documentation
	// for the contents of response.
	fmt.Println(response.SuccessCount, "tokens were subscribed successfully")
	// [END subscribe_golang]
}

func unsubscribeFromTopic(ctx context.Context, client *messaging.Client) {
	topic := "highScores"

	// [START unsubscribe_golang]
	// These registration tokens come from the client FCM SDKs.
	registrationTokens := []string{
		"YOUR_REGISTRATION_TOKEN_1",
		// ...
		"YOUR_REGISTRATION_TOKEN_n",
	}

	// Unsubscribe the devices corresponding to the registration tokens from
	// the topic.
	response, err := client.UnsubscribeFromTopic(ctx, registrationTokens, topic)
	if err != nil {
		log.Fatalln(err)
	}
	// See the TopicManagementResponse reference documentation
	// for the contents of response.
	fmt.Println(response.SuccessCount, "tokens were unsubscribed successfully")
	// [END unsubscribe_golang]
}
