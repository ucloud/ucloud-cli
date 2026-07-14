package message

import (
	"github.com/ucloud/ucloud-cli/internal/common"
	urocketmq "github.com/ucloud/ucloud-sdk-go/services/urocketmq"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// messageRow is the full-field row (json/yaml mode). Mapped from SDK MessageDetail/MessageBaseInfo.
// query-by-key and query-by-topic return MessageBaseInfo (no Body); Body field is empty.
type messageRow struct {
	MsgId     string
	Key       string
	Tag       string
	StoreTime string
	Topic     string
	Body      string
}

// messageRowDefault is the default curated columns in table mode.
type messageRowDefault struct {
	MsgId     string
	Key       string
	Tag       string
	StoreTime string
	Topic     string
}

// formatStoreTime formats a Unix millisecond timestamp as a date-time string.
func formatStoreTime(ms int) string {
	if ms == 0 {
		return ""
	}
	return common.FormatDateTime(ms / 1000)
}

// toMessageDetailRows converts a SDK MessageDetail slice to a messageRow slice.
func toMessageDetailRows(details []urocketmq.MessageDetail) []messageRow {
	rows := make([]messageRow, 0, len(details))
	for _, d := range details {
		rows = append(rows, messageRow{
			MsgId:     d.MsgId,
			Key:       d.Properties.KEYS,
			Tag:       d.Properties.TAGS,
			StoreTime: formatStoreTime(d.StoreTimestamp),
			Topic:     d.Topic,
			Body:      d.MessageBody,
		})
	}
	return rows
}

// toMessageBaseInfoRows converts a SDK MessageBaseInfo slice to a messageRow slice.
func toMessageBaseInfoRows(infos []urocketmq.MessageBaseInfo) []messageRow {
	rows := make([]messageRow, 0, len(infos))
	for _, info := range infos {
		rows = append(rows, messageRow{
			MsgId:     info.MsgId,
			Key:       info.Properties.KEYS,
			Tag:       info.Properties.TAGS,
			StoreTime: formatStoreTime(info.StoreTimestamp),
			Topic:     info.Topic,
		})
	}
	return rows
}

// printMessageRows routes output by ctx.Format(). json/yaml prints full fields, table prints curated columns.
func printMessageRows(ctx *cli.Context, rows []messageRow) {
	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(rows)
		return
	}
	defaultRows := make([]messageRowDefault, 0, len(rows))
	for _, r := range rows {
		defaultRows = append(defaultRows, messageRowDefault{
			MsgId: r.MsgId, Key: r.Key, Tag: r.Tag, StoreTime: r.StoreTime, Topic: r.Topic,
		})
	}
	ctx.PrintList(defaultRows)
}
