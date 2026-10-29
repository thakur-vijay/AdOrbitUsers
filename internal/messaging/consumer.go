package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/adorbit/users/internal/users"
	amqp "github.com/rabbitmq/amqp091-go"
)

const createUserQueue = "users.create"

type CreateUserRequest struct {
	UserID   string `json:"user_id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type CreateUserResponse struct {
	UserID string `json:"user_id"`
	Error  string `json:"error,omitempty"`
}

type Consumer struct {
	rabbitMQ *RabbitMQ
	usersSvc *users.Service
}

func NewConsumer(rabbitMQ *RabbitMQ, usersSvc *users.Service) *Consumer {
	return &Consumer{
		rabbitMQ: rabbitMQ,
		usersSvc: usersSvc,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	channel, err := c.rabbitMQ.Connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	queue, err := channel.QueueDeclare(
		createUserQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	if err := channel.Qos(1, 0, false); err != nil {
		return err
	}

	messages, err := channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("RabbitMQ consumer listening on queue %s", queue.Name)

	for {
		select {
		case <-ctx.Done():
			return nil

		case message, ok := <-messages:
			if !ok {
				return errors.New("RabbitMQ message channel closed")
			}

			c.handleMessage(ctx, channel, message)
		}
	}
}

func (c *Consumer) handleMessage(
	ctx context.Context,
	channel *amqp.Channel,
	message amqp.Delivery,
) {
	var request CreateUserRequest

	response := CreateUserResponse{}

	if err := json.Unmarshal(message.Body, &request); err != nil {
		response.Error = "invalid request payload"
	} else {
		requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		err := c.usersSvc.Create(requestCtx, users.UserDTO{
			ID:        request.UserID,
			FullName:  request.FullName,
			Email:     request.Email,
			Role:      request.Role,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})

		if err != nil {
			log.Printf("Failed to create user %s: %v", request.UserID, err)
			response.Error = "failed to create user"
		} else {
			response.UserID = request.UserID
		}
	}

	body, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal CreateUser response: %v", err)
		_ = message.Nack(false, true)
		return
	}

	if message.ReplyTo != "" {
		err = channel.PublishWithContext(
			ctx,
			"",
			message.ReplyTo,
			false,
			false,
			amqp.Publishing{
				ContentType:   "application/json",
				CorrelationId: message.CorrelationId,
				Body:          body,
			},
		)
		if err != nil {
			log.Printf("Failed to publish CreateUser response: %v", err)
			_ = message.Nack(false, true)
			return
		}
	}

	if err := message.Ack(false); err != nil {
		log.Printf("Failed to acknowledge message: %v", err)
	}
}
