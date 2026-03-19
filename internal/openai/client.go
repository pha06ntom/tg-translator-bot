package openaiwrap

import (
	"context"
	"fmt"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"

	"github.com/pha06ntom/tg-translator-bot/internal/translate"
)

type Client struct {
	cli   openai.Client
	model string
}

func New(apiKey string, modelName string) *Client {
	if modelName == "" {
		modelName = "gpt-5.2"
	}

	return &Client{
		cli:   openai.NewClient(option.WithAPIKey(apiKey)),
		model: modelName,
	}
}

func (c *Client) Translate(ctx context.Context, fromLang, toLang, text string) (string, error) {
	return c.TranslateWithMode(ctx, fromLang, toLang, text, translate.ModeDefault)
}

func (c *Client) TranslateWithMode(ctx context.Context, fromLang, toLang, text string, mode translate.Mode) (string, error) {
	if text == "" {
		return "", nil
	}

	prompt := buildPrompt(fromLang, toLang, text, mode)

	resp, err := c.cli.Responses.New(ctx, responses.ResponseNewParams{
		Model: c.model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(prompt),
		},
	})
	if err != nil {
		return "", err
	}

	return resp.OutputText(), nil
}

func buildPrompt(fromLang, toLang, text string, mode translate.Mode) string {
	if mode == translate.ModeDrawing {
		return fmt.Sprintf(`Ты обрабатываешь OCR-текст с инженерного чертежа.

Твоя задача:
извлечь только осмысленные технические требования, примечания и текстовые указания и перевести их с языка %s на язык %s.

Правила обработки:
- убери мусор OCR: случайные символы, обрывки слов, одиночные буквы, бессмысленные строки
- игнорируй размеры, выноски, рамки, таблицы и штампы, если они не содержат явных текстовых требований
- оставь только технические требования, примечания и текстовые указания
- если строка не имеет смысла, пропусти её
- исправляй очевидные OCR-ошибки, если исправление очевидно по контексту

Ограничения:
- не изменяй числовые значения
- не изменяй размеры
- не изменяй обозначения стандартов (GOST, ГОСТ, ISO и т.п.)
- не изменяй марки материалов
- не изменяй коды, позиции, артикулы, номера
- не придумывай текст, которого нет в OCR

Формат результата:
- верни аккуратный нумерованный список
- каждая строка — отдельное требование
- без пояснений от себя
- если удалось распознать только 1 осмысленное требование, верни только его
- если осмысленных требований не найдено, верни: "Не удалось надёжно извлечь технические требования с чертежа."

OCR-текст:
%s`, fromLang, toLang, text)
	}

	return fmt.Sprintf(`Ты профессиональный переводчик.
Переведи текст с языка %s на язык %s.

Требования:
- сохрани смысл и стиль
- не добавляй пояснений
- не комментируй
- если есть списки и таблицы, по возможности сохрани форматирование

Текст:
%s`, fromLang, toLang, text)
}
