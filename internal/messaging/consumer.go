package messaging

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/adorbit/users/internal/users"
	"github.com/jackc/pgx/v5/pgconn"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	createUserQueue = "users.create"
	fetchUserQueue  = "users.fetch"

	userAlreadyExists  = "USER_ALREADY_EXISTS"
	userCreationFailed = "USER_CREATION_FAILED"
	userNotFound       = "USER_NOT_FOUND"
	userFetchFailed    = "USER_FETCH_FAILED"
	invalidRequest     = "INVALID_REQUEST"
)

type CreateUserRequest struct {
	UserID   string `json:"user_id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type FetchUserRequest struct {
	UserID string `json:"user_id"`
}

type UserResponse struct {
	UserID    string    `json:"user_id,omitempty"`
	FullName  string    `json:"full_name,omitempty"`
	Email     string    `json:"email,omitempty"`
	Role      string    `json:"role,omitempty"`
	Code      string    `json:"code,omitempty"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Consumer struct {
	rabbitMQ *RabbitMQ
	usersSvc *users.Service
}

func NewConsumer(
	rabbitMQ *RabbitMQ,
	usersSvc *users.Service,
) *Consumer {
	return &Consumer{
		rabbitMQ: rabbitMQ,
		usersSvc: usersSvc,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	if c.rabbitMQ == nil || c.rabbitMQ.Connection == nil {
		return errors.New("RabbitMQ connection is not initialized")
	}

	channel, err := c.rabbitMQ.Connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	createQueue, err := channel.QueueDeclare(
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

	fetchQueue, err := channel.QueueDeclare(
		fetchUserQueue,
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

	createMessages, err := channel.Consume(
		createQueue.Name,
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

	fetchMessages, err := channel.Consume(
		fetchQueue.Name,
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

	log.Printf(
		"RabbitMQ consumers listening on queues: %s, %s",
		createQueue.Name,
		fetchQueue.Name,
	)

	for {
		select {
		case <-ctx.Done():
			return nil

		case message, ok := <-createMessages:
			if !ok {
				return errors.New("Create user message channel closed")
			}

			c.handleMessage(ctx, channel, message)

		case message, ok := <-fetchMessages:
			if !ok {
				return errors.New("Fetch user message channel closed")
			}

			c.handleFetchMessage(ctx, channel, message)
		}
	}
}

func (c *Consumer) handleMessage(
	ctx context.Context,
	channel *amqp.Channel,
	message amqp.Delivery,
) {
	var request CreateUserRequest

	response := UserResponse{}

	if err := json.Unmarshal(message.Body, &request); err != nil {
		response.Code = invalidRequest
		response.Message = "Invalid request payload"

	} else if request.UserID == "" {
		response.Code = invalidRequest
		response.Message = "User ID is required"

	} else {
		requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

		now := time.Now().UTC()

		user, err := c.usersSvc.Create(
			requestCtx,
			users.UserDTO{
				ID:        request.UserID,
				FullName:  request.FullName,
				Email:     request.Email,
				Role:      request.Role,
				CreatedAt: now,
				UpdatedAt: now,
			},
		)

		if err != nil {
			// Existing error handling
		} else {
			response = UserResponse{
				UserID:    user.ID,
				FullName:  user.FullName,
				Email:     user.Email,
				Role:      user.Role,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			}
		}

		cancel()

		if err != nil {
			log.Printf(
				"Failed to create user %s: %v",
				request.UserID,
				err,
			)

			response.Code = userCreationFailed
			response.Message = "Failed to create user"

			var pgErr *pgconn.PgError

			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					response.Code = userAlreadyExists
					response.Message = "User already exists"
				}
			}

		} else {
			response.UserID = request.UserID
		}
	}

	c.publishResponse(ctx, channel, message, response)
}

func (c *Consumer) handleFetchMessage(
	ctx context.Context,
	channel *amqp.Channel,
	message amqp.Delivery,
) {
	var request FetchUserRequest

	response := UserResponse{}

	if err := json.Unmarshal(message.Body, &request); err != nil {
		response.Code = invalidRequest
		response.Message = "Invalid request payload"

	} else if request.UserID == "" {
		response.Code = invalidRequest
		response.Message = "User ID is required"

	} else {
		requestCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		user, err := c.usersSvc.FetchByID(
			requestCtx,
			request.UserID,
		)

		if err != nil {
			log.Printf(
				"Failed to fetch user %s: %v",
				request.UserID,
				err,
			)

			response.Code = userFetchFailed
			response.Message = "Failed to fetch user"

			if errors.Is(err, sql.ErrNoRows) {
				response.Code = userNotFound
				response.Message = "User not found"
			}

		} else if user == nil {
			response.Code = userNotFound
			response.Message = "User not found"

		} else {
			response = UserResponse{
				UserID:    user.ID,
				FullName:  user.FullName,
				Email:     user.Email,
				Role:      user.Role,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
			}
		}
	}

	c.publishResponse(ctx, channel, message, response)
}

func (c *Consumer) publishResponse(
	ctx context.Context,
	channel *amqp.Channel,
	message amqp.Delivery,
	response UserResponse,
) {
	body, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal RabbitMQ response: %v", err)
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
			log.Printf("Failed to publish RabbitMQ response: %v", err)
			_ = message.Nack(false, true)
			return
		}
	}

	if err := message.Ack(false); err != nil {
		log.Printf("Failed to acknowledge RabbitMQ message: %v", err)
	}
}
