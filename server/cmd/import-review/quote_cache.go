package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// One rolling window per stock. Historical review publications stay in MarketReviewDay.
type reviewQuoteWindow struct {
	Symbol string `gorm:"primaryKey;size:12"`
	Data   []byte `gorm:"type:mediumblob;not null"`
}

var reviewStock = regexp.MustCompile(`^(6[0-9]{5}\.SH|(00|30)[0-9]{4}\.SZ)$`)

func quoteCache(db *gorm.DB, readDir string, paths []string) error {
	if err := db.AutoMigrate(&reviewQuoteWindow{}); err != nil {
		return err
	}
	if readDir != "" {
		rows, err := db.Model(&reviewQuoteWindow{}).Select("symbol, data").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var row reviewQuoteWindow
			if err := rows.Scan(&row.Symbol, &row.Data); err != nil {
				return err
			}
			if !reviewStock.MatchString(row.Symbol) {
				return fmt.Errorf("invalid quote cache symbol")
			}
			if err := os.WriteFile(filepath.Join(readDir, row.Symbol+".gz"), row.Data, 0600); err != nil {
				return err
			}
		}
		return rows.Err()
	}
	batch := make([]reviewQuoteWindow, 0, len(paths))
	for _, path := range paths {
		name := filepath.Base(path)
		symbol := name[:len(name)-len(filepath.Ext(name))]
		if !reviewStock.MatchString(symbol) || filepath.Ext(name) != ".gz" {
			return fmt.Errorf("invalid quote window file")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		batch = append(batch, reviewQuoteWindow{Symbol: symbol, Data: data})
	}
	return db.Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(batch, 32).Error
}
