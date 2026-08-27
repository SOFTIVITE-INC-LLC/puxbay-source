package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User represents an authentication user (replaces Django's auth.User).
type User struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username         string         `gorm:"size:150;uniqueIndex;not null" json:"username"`
	Email            string         `gorm:"size:254;uniqueIndex;not null" json:"email"`
	Password         string         `gorm:"size:255;not null" json:"-"` // bcrypt hash, never exposed
	FirstName        string         `gorm:"size:150" json:"first_name"`
	LastName         string         `gorm:"size:150" json:"last_name"`
	Phone                 string         `gorm:"size:20" json:"phone"`
	IsActive              bool           `gorm:"default:true" json:"is_active"`
	IsSuperuser           bool           `gorm:"default:false" json:"is_superuser"`
	IsStaff               bool           `gorm:"default:false" json:"is_staff"`
	IsEmailVerified       bool           `gorm:"default:false" json:"is_email_verified"`
	EmailVerificationToken *string       `gorm:"size:100;index" json:"-"`
	EmailVerificationExpiry *time.Time   `json:"-"`
	EmailVerificationCode *string        `gorm:"size:10;index" json:"-"`
	LastLogin             *time.Time     `json:"last_login,omitempty"`
	DateJoined            time.Time      `gorm:"autoCreateTime" json:"date_joined"`
	TokenVersion          int            `gorm:"default:1" json:"-"`
	ResetToken            *string        `gorm:"size:100;index" json:"-"`
	ResetTokenExpiry      *time.Time     `json:"-"`
	RequirePasswordChange bool           `gorm:"default:false" json:"require_password_change"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Profiles []UserProfile `gorm:"foreignKey:UserID" json:"profiles,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (User) TableName() string {
	return "public.users"
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	if u.FirstName != "" || u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.Username
}

// UserProfile links a user to a specific tenant/branch with a role.
// Maps from Django: accounts.models.UserProfile
type UserProfile struct {
	ID       uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TenantID uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	BranchID *uuid.UUID `gorm:"type:uuid;index" json:"branch_id,omitempty"`
	Branch   *Branch    `gorm:"-" json:"branch,omitempty"`
	RoleID   uuid.UUID  `gorm:"type:uuid;index" json:"role_id"`
	Role     *Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`

	// Permissions
	CanPerformCreditSales bool `gorm:"default:false" json:"can_perform_credit_sales"`

	// Payroll & Compensation
	BaseSalary    float64        `gorm:"type:decimal(12,2);default:0" json:"base_salary"`
	HourlyRate    float64        `gorm:"type:decimal(10,2);default:0" json:"hourly_rate"`
	PaymentMethod string         `gorm:"size:20;default:'cash'" json:"payment_method"`
	BankDetails   datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"bank_details"`

	// Security
	Is2FAEnabled    bool    `gorm:"column:is_2fa_enabled;default:false" json:"is_2fa_enabled"`
	OTPSecret       *string `gorm:"column:otp_secret;type:text" json:"-"` // Encrypted
	POSPin          *string `gorm:"type:text" json:"-"`                   // Encrypted
	IsEmailVerified bool    `gorm:"default:false" json:"is_email_verified"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"-"`
}

func (up *UserProfile) BeforeCreate(tx *gorm.DB) error {
	if up.ID == uuid.Nil {
		up.ID = uuid.New()
	}
	return nil
}

func (UserProfile) TableName() string {
	return "public.user_profiles"
}

// ValidRoles (Deprecated, handled by RBAC dynamic engine)
func ValidRoles() []string {
	return []string{"admin", "manager", "branch_manager", "procurement_manager", "sales", "financial", "supplier", "cashier"}
}
