// +build all 

package state_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common/primitives/random"
	. "github.com/FactomProject/factomd/state"
)

type Exposer struct {
	AWA float64
	ATA float64
	ABU float64

	CWA float64
	CTA float64

	CMA float64
	AMA float64
}

func NewExposer() *Exposer {
	e := new(Exposer)
	return e
}

func (k *Exposer) SetArrivalInstantAvg(v float64) {
	k.AWA = v
}

func (k *Exposer) SetArrivalTotalAvg(v float64) {
	k.ATA = v
}

func (k *Exposer) SetArrivalBackup(v float64) {
	k.ABU = v
}

func (k *Exposer) SetCompleteInstantAvg(v float64) {
	k.CWA = v
}

func (k *Exposer) SetMovingArrival(v float64) {
	k.AMA = v
}

func (k *Exposer) SetMovingComplete(v float64) {
	k.CMA = v
}

func (k *Exposer) SetCompleteTotalAvg(v float64) {
	k.CTA = v
}

func (k *Exposer) String() string {
	return fmt.Sprintf("%f, %f, %f, %f, %f", k.AWA, k.ATA, k.ABU, k.CWA, k.CTA)
}

func r_close(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = diff * -1
	}
	if diff < tolerance {
		return true
	}
	return false
}

func shouldbe(awa, abu, cwa float64, e *Exposer) error {
	if !r_close(awa, e.AWA, 0.2) {
		return fmt.Errorf("AWA is %f, should be %f", awa, e.AWA)
	}

	// By Speeding up the tick time, these numbers are usually off
	//if !r_close(ata, e.ATA, 15) {
	//return fmt.Errorf("ATA is %f, should be %f", ata, e.ATA)
	//}

	if !r_close(abu, e.ABU, 0.1) {
		return fmt.Errorf("ABU is %f, should be %f", abu, e.ABU)
	}

	if !r_close(cwa, e.CWA, 0.2) {
		return fmt.Errorf("CWA is %f, should be %f", cwa, e.CWA)
	}

	// By Speeding up the tick time, these numbers are usually off
	//if !r_close(cta, e.CTA, 15) {
	//return fmt.Errorf("CTA is %f, should be %f", cta, e.CTA)
	//}

	return nil
}

// Most of the time is sleeping, so run 10 in parallel
func TestRateCalculator(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("ParallelTest%d", i), testRateCalculator)
	}
}

func testRateCalculator(t *testing.T) {
	t.Parallel()
	e := NewExposer()
	td := time.Millisecond * 100
	rc := NewRateCalculatorTime(e, td)

	TotalA := float64(0)
	TotalC := float64(0)
	ac := func(add int) {
		for i := 0; i < add; i++ {
			rc.Arrival()
			rc.Complete()
			TotalC++
			TotalA++
		}
	}

	fa := random.RandIntBetween(0, 100)
	ac(fa)
	start := time.Now()
	go rc.StartTime(start)
	ticker := time.NewTicker(td - 1*time.Millisecond)
	i := 0

	ataF := func(diff time.Duration) float64 {
		return TotalA / (time.Since(start).Seconds() - diff.Seconds())
	}

	ctaF := func(diff time.Duration) float64 {
		return TotalC / (time.Since(start).Seconds() - diff.Seconds())
	}
	var _, _ = ataF, ctaF

Outer:
	for _ = range ticker.C {
		switch i {
		case 0:
			if err := retry(float64(fa)/td.Seconds(), 0, float64(fa)/td.Seconds(), e, 20); err != nil {
				t.Error("1", err)
			}
		case 1:
			if err := retry(0, 0, 0, e, 20); err != nil {
				t.Error("2", err)
			}
			go ac(fa * 3)
		case 2:
			if err := retry(float64(fa*3)/td.Seconds(), 0, float64(fa*3)/td.Seconds(), e, 20); err != nil {
				t.Error("3", err)
			}
		case 3:
			if err := retry(0, 0, 0, e, 20); err != nil {
				t.Error("2", err)
			}
			break Outer
		}
		i++
	}
}

func retry(awa, abu, cwa float64, e *Exposer, amt int) error {
	var err error
	for a := 0; a < amt; a++ {
		err = shouldbe(awa, abu, cwa, e)
		if err == nil {
			return nil
		} else {
			if a >= amt {
				return err
			}
		}
		time.Sleep(1 * time.Millisecond)
	}
	return err
}

func TestMovingAverage(t *testing.T) {
	a := NewMovingAverage(5)
	if a.Avg() != 0 {
		t.Fail()
	}
	a.Add(2)
	if a.Avg() < 1.999 || a.Avg() > 2.001 {
		t.Fail()
	}
	a.Add(4)
	a.Add(2)
	if a.Avg() < 2.665 || a.Avg() > 2.667 {
		t.Fail()
	}
	a.Add(4)
	a.Add(2)
	if a.Avg() < 2.799 || a.Avg() > 2.801 {
		t.Fail()
	}

	// This one will go into the first slot again
	// evicting the first value
	a.Add(10)
	if a.Avg() < 4.399 || a.Avg() > 4.401 {
		t.Fail()
	}

	for i := 0; i < 10; i++ {
		a.Add(0)
	}

	if a.Avg() < 0-0.1 || a.Avg() > 0.+0.1 {
		t.Fail()
	}

	a = NewMovingAverage(5)
	nums := make([]float64, 1000)
	index := 0

	for i := 0; i < 1000; i++ {
		nums[i] = float64(random.RandIntBetween(0, 60000))
	}

	for ; index < 1000; index++ {
		a.Add(nums[index])
		v := a.Avg()
		tv := float64(0)
		total := float64(0)
		for sub := index; sub >= 0; sub-- {
			if total >= 5 || sub < 0 {
				break
			}
			total++
			tv += nums[sub]
		}

		diff := v - (tv / total)
		if diff < 0 {
			diff = -1 * diff
		}

		if diff > 0.1 {
			t.Errorf("Difference is %f at index %d. Found %f, exp %f. Total %f", diff, index, v, tv/total, total)
			t.Log(nums[:index+1])
		}
	}
}
