package Tables

import "github.com/grdvsng/stock-analysis/Storage"

func Create(db *Storage.QuestDB) error {
	for _, model := range []Storage.Model{Stock{}, Contract{}} {
		if e := db.Create(model); e != nil {
			return e
		}
	}

	return nil
}
