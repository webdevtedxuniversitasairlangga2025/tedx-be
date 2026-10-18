package service

import (
	"testing"

	"github.com/webdevtedxuniversitasairlangga/modules/order/dto"
)

func TestGroupTicketsByEmail(t *testing.T) {
	ticket := func(name, email string) dto.AttendeeTicketResponse {
		return dto.AttendeeTicketResponse{ID: name, AttendeeName: name, AttendeeEmail: email}
	}

	t.Run("tiga email berbeda → tiga grup urut first-seen", func(t *testing.T) {
		groups := groupTicketsByEmail([]dto.AttendeeTicketResponse{
			ticket("Andi", "andi@gmail.com"),
			ticket("Budi", "budi@gmail.com"),
			ticket("Cici", "cici@gmail.com"),
		}, "buyer@gmail.com")
		if len(groups) != 3 {
			t.Fatalf("groups = %d, mau 3", len(groups))
		}
		if groups[0].Email != "andi@gmail.com" || groups[1].Email != "budi@gmail.com" || groups[2].Email != "cici@gmail.com" {
			t.Fatalf("urutan email tidak sesuai: %+v", groups)
		}
	})

	t.Run("email sama (beda kapital) → digabung satu grup", func(t *testing.T) {
		groups := groupTicketsByEmail([]dto.AttendeeTicketResponse{
			ticket("Andi", "Andi@Gmail.com"),
			ticket("Andi 2", "andi@gmail.COM"),
		}, "")
		if len(groups) != 1 || len(groups[0].Tickets) != 2 {
			t.Fatalf("groups = %+v, mau 1 grup berisi 2 tiket", groups)
		}
	})

	t.Run("email kosong → fallback pembeli", func(t *testing.T) {
		groups := groupTicketsByEmail([]dto.AttendeeTicketResponse{
			ticket("Andi", ""),
		}, " buyer@gmail.com ")
		if len(groups) != 1 || groups[0].Email != "buyer@gmail.com" {
			t.Fatalf("groups = %+v, mau fallback buyer@gmail.com", groups)
		}
	})

	t.Run("email kosong tanpa fallback → dilewatkan", func(t *testing.T) {
		groups := groupTicketsByEmail([]dto.AttendeeTicketResponse{
			ticket("Andi", ""),
		}, "")
		if len(groups) != 0 {
			t.Fatalf("groups = %+v, mau kosong", groups)
		}
	})
}
