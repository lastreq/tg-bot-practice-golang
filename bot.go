package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
)

const (
	BotToken   = "5063425452:AAHSgIlzli2FQnnfnZqfVqItLoPVedHEQuo"
	WebhookURL = "https://tg-bot-golang-practice.herokuapp.com"
)

func getSchedule(group string, day int) string {

	var resultFull string
	url := fmt.Sprintf("https://itmo.ru/ru/schedule/0/%s/raspisanie_zanyatiy_%s.htm", group, group)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("status code error: %d %s\n", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		log.Println(err)
	}
	tableContent := doc.Find("table")
	dayString := fmt.Sprintf("%vday", day)
	doc.Find("table").Each(func(index int, item *goquery.Selection) {
		if item.AttrOr("id", "") == dayString {
			resultFull += parseDay(item)
		}
	})

	if len(tableContent.Nodes) == 0 {
		return fmt.Sprintf("Расписание занятий для группы %v не найдено", group)
	}
	if day == 7 {
		return fmt.Sprintf("Никакой учебы по воскресеньям")
	}
	//doc1 := goquery.NewDocumentFromNode(tableContent.Nodes[day])
	resultFull += getToday(doc)

	return resultFull
}
func getToday(doc *goquery.Document) string {
	var day string
	doc.Find("h2").Each(func(index int, item *goquery.Selection) {
		if item.AttrOr("class", "") == "schedule-week" {
			day = "Сегодня: " + item.Text() + "\n"

		}
	})
	return day
}

func parseDay(doc1 *goquery.Selection) string {
	var timeSlice, locationSlice, lessonsSlice, roomSlice, teacherSlice, lessonsTypeSlice, weeksSlice []string
	var result string
	doc1.Find("td").Each(func(index int, item *goquery.Selection) {

		if item.AttrOr("class", "") == "time" {
			textTime := item.Find("span").First().Text()
			textWeeks := item.Find("div").Text()
			timeSlice = append(timeSlice, textTime)
			weeksSlice = append(weeksSlice, textWeeks)
		}

		if item.AttrOr("class", "") == "room" {
			textRoom := item.Find("dd").Text()
			textLoc := item.Find("span").Text()
			locationSlice = append(locationSlice, textLoc)
			roomSlice = append(roomSlice, textRoom)
		}

		if item.AttrOr("class", "") == "lesson" {
			textLesson := item.Find("dd").Text()
			lessonsSlice = append(lessonsSlice, strings.TrimSpace(textLesson))
			textTeacher := item.Find("b").Text()
			teacherSlice = append(teacherSlice, strings.TrimSpace(textTeacher))
		}
		if item.AttrOr("class", "") == "lesson-format" {
			textLessonsType := item.Text()
			lessonsTypeSlice = append(lessonsTypeSlice, strings.TrimSpace(textLessonsType))

		}

	})
	for i := 0; i < len(lessonsSlice); i++ {
		result += timeSlice[i] + "  " + locationSlice[i] + "  " + roomSlice[i] + "  " + lessonsSlice[i] + "  " + teacherSlice[i] + "  " + lessonsTypeSlice[i] + "  Недели:" + weeksSlice[i] + "\n"
	}
	return result
}

func getDay(s string) int {
	var dayNum int
	switch {
	case strings.Contains(s, "понедельник") || strings.Contains(s, "Понедельник"):
		dayNum = 1
	case strings.Contains(s, "вторник") || strings.Contains(s, "Вторник"):
		dayNum = 2
	case strings.Contains(s, "среда") || strings.Contains(s, "Среда"):
		dayNum = 3
	case strings.Contains(s, "четверг") || strings.Contains(s, "Четверг"):
		dayNum = 4
	case strings.Contains(s, "пятница") || strings.Contains(s, "Пятница"):
		dayNum = 5
	case strings.Contains(s, "суббота") || strings.Contains(s, "Суббота"):
		dayNum = 6
	case strings.Contains(s, "воскресенье") || strings.Contains(s, "Воскресенье"):
		dayNum = 7

	}

	return dayNum
}
func formatInputString(s string) string {
	stringToTrim := s
	stringToTrim = strings.TrimSpace(stringToTrim)
	r := regexp.MustCompile("\\s+")
	replace := r.ReplaceAllString(stringToTrim, " ")

	result := strings.Split(strings.ToUpper(replace), " ")
	return result[1]
}

func main() {
	telegramBot()
}
func telegramBot() {
	var startReg = regexp.MustCompile(`/start`)
	var dayReg = regexp.MustCompile(`[пП]онедельник |[вВ]торник |[сС]реда |[чЧ]етверг |[пП]ятница |[сС]уббота |[вВ]оскресенье `)
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		panic(err)
	}

	updates := bot.ListenForWebhook("/")

	go http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	fmt.Println("start listen:" + os.Getenv("PORT"))

	// получаем все обновления из канала updates
	for update := range updates {
		if reflect.TypeOf(update.Message.Text).Kind() == reflect.String && update.Message.Text != "" {

			switch {
			case startReg.MatchString(update.Message.Text):
				//Отправлем сообщение
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет, этот бот предназначен для получения расписания с сайта итмо: https://itmo.ru/ru//schedule/raspisanie_zanyatiy.htm\nДля получения расписания ввиде: \"день_недели группа\"\nПример: вторник D3110")
				bot.Send(msg)

			case dayReg.MatchString(update.Message.Text):
				group := update.Message.Text
				var result string
				var day int
				day = getDay(group)

				result = getSchedule(formatInputString(group), day)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, result)
				bot.Send(msg)

			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный ввод\nПример: Вторник D3110")
				bot.Send(msg)
			}

		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неверный ввод, это нет текст")
			bot.Send(msg)
		}

	}
}
