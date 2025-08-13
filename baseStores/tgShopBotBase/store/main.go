/* THIS HANDLES EVERYTHING THAT THE FRONT END
OF A STORE REQUIRES
Categories() returns shop catergories
Listings() returns listing of a category
Item() returns the item listing from category */

package store

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Catergories() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("Hats", "category"),
			tgbotapi.NewInlineKeyboardButtonData("Coats", "category"),
			tgbotapi.NewInlineKeyboardButtonData("Main Menu", "back_main"),
		},
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	return keyboard
}

// fun Listing() will handle api connection for database queries specific to vender_listings
func Listings() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("Hat", "item")},
		tgbotapi.NewInlineKeyboardRow(
			//tgbotapi.NewInlineKeyboardButtonData("Back", "back"),
			tgbotapi.NewInlineKeyboardButtonData("Main Menu", "back_main"),
		),
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	return keyboard
}

func Item() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("Quantity +", "quantity+"),
			tgbotapi.NewInlineKeyboardButtonData(" {.value}", "quantity"),
			tgbotapi.NewInlineKeyboardButtonData("Quantity -", "quantity-"),
		},
		{tgbotapi.NewInlineKeyboardButtonData("Add to Basket", "basket_add")},
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Back", "back"),
			tgbotapi.NewInlineKeyboardButtonData("Main Menu", "back_main"),
		),
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	return keyboard
}
