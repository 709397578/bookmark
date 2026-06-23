package models

import (
	"crypto/rand"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        string    `gorm:"type:uuid;primary_key" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	Role      string    `gorm:"default:user" json:"role"` // admin, user
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Collection 收藏夹模型
type Collection struct {
	ID            string    `gorm:"type:uuid;primary_key" json:"id"`
	Name          string    `gorm:"not null" json:"name"`
	Slug          string    `gorm:"uniqueIndex" json:"slug"`
	Description   string    `json:"description"`
	Icon          string    `json:"icon"`
	IsPublic      bool      `json:"isPublic"`
	SortOrder     int       `gorm:"default:0" json:"sortOrder"`
	UserID        string    `gorm:"type:uuid" json:"userId"`
	User          User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BookmarkCount int       `gorm:"-" json:"bookmarkCount"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Folder 文件夹模型
type Folder struct {
	ID            string     `gorm:"type:uuid;primary_key" json:"id"`
	Name          string     `gorm:"not null" json:"name"`
	Icon          string     `json:"icon"`
	Color         string     `json:"color"`
	CollectionID  string     `gorm:"type:uuid;not null" json:"collectionId"`
	Collection    Collection `gorm:"foreignKey:CollectionID" json:"collection,omitempty"`
	ParentID      *string    `gorm:"type:uuid" json:"parentId"`
	SortOrder     int        `gorm:"default:0" json:"sortOrder"`
	BookmarkCount int        `gorm:"-" json:"bookmarkCount"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// Bookmark 书签模型
type Bookmark struct {
	ID           string    `gorm:"type:uuid;primary_key" json:"id"`
	Title        string    `gorm:"not null" json:"title"`
	URL          string    `gorm:"not null" json:"url"`
	Description  string    `json:"description"`
	Icon         string    `json:"icon"`
	IsFeatured   bool      `gorm:"default:false" json:"isFeatured"`
	SortOrder    int       `gorm:"default:0" json:"sortOrder"`
	CollectionID string    `gorm:"type:uuid;not null" json:"collectionId"`
	FolderID     *string   `gorm:"type:uuid" json:"folderId"`
	Collection   Collection `gorm:"foreignKey:CollectionID" json:"collection,omitempty"`
	Folder       *Folder   `gorm:"foreignKey:FolderID" json:"folder,omitempty"`
	Tags         []Tag     `gorm:"many2many:bookmark_tags;" json:"tags,omitempty"`
	HasSnapshot  bool      `gorm:"default:false" json:"hasSnapshot"`
	SnapshotURL  string    `json:"snapshotUrl"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Tag 标签模型
type Tag struct {
	ID        string     `gorm:"type:uuid;primary_key" json:"id"`
	Name      string     `gorm:"uniqueIndex;not null" json:"name"`
	Bookmarks []Bookmark `gorm:"many2many:bookmark_tags;" json:"bookmarks,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// Setting 设置模型
type Setting struct {
	ID        string    `gorm:"type:uuid;primary_key" json:"id"`
	Key       string    `gorm:"uniqueIndex;not null" json:"key"`
	Value     string    `json:"value"`
	Type      string    `gorm:"default:string" json:"type"` // string, number, boolean, json
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Image 图片模型
type Image struct {
	ID        string    `gorm:"type:uuid;primary_key" json:"id"`
	URL       string    `gorm:"not null" json:"url"`
	Filename  string    `gorm:"not null" json:"filename"`
	Size      int64     `json:"size"`
	MimeType  string    `json:"mimeType"`
	Key       string    `json:"key"` // 用于标识图片用途
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func generateUUID() string {
	id := make([]byte, 16)
	rand.Read(id)
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		id[0:4], id[4:6], id[6:8], id[8:10], id[10:])
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == "" {
		u.ID = generateUUID()
	}
	return nil
}

func (c *Collection) BeforeCreate(_ *gorm.DB) error {
	if c.ID == "" {
		c.ID = generateUUID()
	}
	return nil
}

func (f *Folder) BeforeCreate(_ *gorm.DB) error {
	if f.ID == "" {
		f.ID = generateUUID()
	}
	return nil
}

func (b *Bookmark) BeforeCreate(_ *gorm.DB) error {
	if b.ID == "" {
		b.ID = generateUUID()
	}
	return nil
}

func (t *Tag) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = generateUUID()
	}
	return nil
}

func (s *Setting) BeforeCreate(_ *gorm.DB) error {
	if s.ID == "" {
		s.ID = generateUUID()
	}
	return nil
}

func (i *Image) BeforeCreate(_ *gorm.DB) error {
	if i.ID == "" {
		i.ID = generateUUID()
	}
	return nil
}

func (User) TableName() string {
	return "users"
}

func (Collection) TableName() string {
	return "collections"
}

func (Folder) TableName() string {
	return "folders"
}

func (Bookmark) TableName() string {
	return "bookmarks"
}

func (Tag) TableName() string {
	return "tags"
}

func (Setting) TableName() string {
	return "settings"
}

func (Image) TableName() string {
	return "images"
}
