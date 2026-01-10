package plugins

import (
	"time"

	"github.com/nxo/engine/internal/models"
	"gorm.io/gorm/clause"
)

// SyncDiscovered upserts all compiled-in plugins into the plugins table.
// It does not change the enabled flag for existing rows.
func SyncDiscovered(deps Deps) error {
	for _, p := range All() {
		pluginRow := &models.Plugin{
			Name:        p.Key(),
			Version:     p.Version(),
			Description: p.Description(),
			Manifest:    models.JSON(p.Manifest()),
			InstalledAt: time.Now(),
		}

		err := deps.DB.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "name"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"version",
					"description",
					"manifest",
					"updated_at",
				}),
			}).
			Create(pluginRow).Error
		if err != nil {
			return err
		}
	}
	return nil
}
