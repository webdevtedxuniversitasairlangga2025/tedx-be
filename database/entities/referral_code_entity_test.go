package entities

import (
	"reflect"
	"testing"
)

func TestReferralCodeUsesUniqueCodeAndQuotaFields(t *testing.T) {
	typeOf := reflect.TypeOf(ReferralCode{})
	for _, fieldName := range []string{"Code", "DiscountPerTicket", "Quota", "QuotaHeld", "QuotaUsed", "IsActive"} {
		if _, ok := typeOf.FieldByName(fieldName); !ok {
			t.Errorf("ReferralCode missing %s field", fieldName)
		}
	}
	code, _ := typeOf.FieldByName("Code")
	if !reflect.ValueOf(code.Tag.Get("gorm")).IsValid() || code.Tag.Get("gorm") == "" {
		t.Fatal("ReferralCode.Code requires a GORM constraint tag")
	}
	if _, ok := typeOf.FieldByName("DeletedAt"); !ok {
		t.Fatal("ReferralCode must retain archived codes with GORM soft delete")
	}
}

func TestOrderStoresReferralDiscountSnapshot(t *testing.T) {
	typeOf := reflect.TypeOf(Order{})
	for _, fieldName := range []string{"ReferralCodeID", "ReferralCodeSnapshot", "ReferralDiscountedCount", "ReferralTotalDiscount"} {
		if _, ok := typeOf.FieldByName(fieldName); !ok {
			t.Errorf("Order missing %s field", fieldName)
		}
	}
}
