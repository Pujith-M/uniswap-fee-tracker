// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/transactions/{txHash}": {
            "get": {
                "description": "Get the transaction fee in USDT for a specific Uniswap WETH-USDC transaction",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "transactions"
                ],
                "summary": "Get transaction fee in USDT",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Transaction Hash",
                        "name": "txHash",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.TransactionResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/models.ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/models.ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "models.ErrorResponse": {
            "description": "Error response when the API request fails",
            "type": "object",
            "properties": {
                "error": {
                    "description": "Error message\n@Description Description of what went wrong",
                    "type": "string"
                }
            }
        },
        "models.TransactionResponse": {
            "description": "Response containing transaction details including gas fees",
            "type": "object",
            "properties": {
                "block_number": {
                    "description": "Block number\n@Description The block number in which this transaction was included",
                    "type": "integer"
                },
                "eth_price": {
                    "description": "ETH price in USDT\n@Description ETH/USDT price at transaction time",
                    "type": "string"
                },
                "fee_eth": {
                    "description": "Fee in ETH\n@Description Transaction fee in ETH",
                    "type": "string"
                },
                "fee_usdt": {
                    "description": "Fee in USDT\n@Description Transaction fee converted to USDT",
                    "type": "string"
                },
                "gas_price": {
                    "description": "Gas price\n@Description Price per unit of gas in Wei",
                    "type": "string"
                },
                "gas_used": {
                    "description": "Gas used\n@Description Amount of gas used by this transaction",
                    "type": "string"
                },
                "status": {
                    "description": "Transaction status\n@Description Current processing status of the transaction",
                    "allOf": [
                        {
                            "$ref": "#/definitions/syncer.TransactionStatus"
                        }
                    ]
                },
                "timestamp": {
                    "description": "Transaction timestamp\n@Description When this transaction was processed",
                    "type": "string"
                },
                "tx_hash": {
                    "description": "Transaction hash\n@Description Unique identifier of the transaction",
                    "type": "string"
                }
            }
        },
        "syncer.TransactionStatus": {
            "type": "string",
            "enum": [
                "PROCESSED",
                "PENDING_PRICE",
                "FAILED"
            ],
            "x-enum-varnames": [
                "StatusProcessed",
                "StatusPendingPrice",
                "StatusFailed"
            ]
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
