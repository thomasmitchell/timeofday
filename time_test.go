package timeofday_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/thomasmitchell/timeofday"
)

var _ = Describe("TimeOfDay", func() {
	Describe("parsing from strings", func() {
		var (
			tod       *TimeOfDay
			err       error
			inputSpec string
			inputLoc  *time.Location
		)
		BeforeEach(func() {
			inputLoc = time.UTC
		})

		JustBeforeEach(func() {
			tod, err = NewFromString(inputSpec, inputLoc)
		})

		Context("when the timespec is bad", func() {
			badSpec := func(desc, spec string) {
				Context(fmt.Sprintf("because %s (%s)", desc, spec), func() {
					BeforeEach(func() {
						inputSpec = spec
					})

					It("should have erred", func() {
						Expect(err).To(HaveOccurred())
					})

					Specify("the returned TimeOfDay value should be nil", func() {
						Expect(tod).To(BeNil())
					})
				})
			}

			badSpec("the input was empty", "")
			badSpec("there was no hour in the input", ":59")
			badSpec("there was a colon but no minute", "03:")
			badSpec("there was a colon but no minute, but there was a meridiem", "03:PM")
			badSpec("there are too many leading zeroes in the hour", "003:59")
			badSpec("there too many leading zeroes in the minute", "03:009")
			badSpec("there aren't enough digits in the minute", "03:9")
			badSpec("the meridiem is fake", "03:09 FM")
			badSpec("the meridiem is missing the `M'", "03:09 A")
			badSpec("the meridiem is only an `M'", "03:09 M")
			badSpec("there is no colon between the hour and minute", "0309 PM")
			badSpec("the input is the word `beep'", "beep")
			badSpec("the input is the word `beep' followed by a valid meridiem", "beep PM")
			badSpec("the meridiem is before the time", "PM 3:09")
			badSpec("the hour is negative", "-3:09")
			badSpec("the minute is negative", "3:-09")
			badSpec("the hour is a higher number than 24", "25:09")
			badSpec("the minute is a higher number than 59", "3:60")
			badSpec("the hour is over 12 and the meridiem is PM", "13:09 PM")
			badSpec("the hour is slightly negative and the meridiem is PM", "-1:00 PM")
			badSpec("the input is a number followed by the word `beep'", "8beep")
			badSpec("the input is a number followed by a colon followed by the word beep", "8:beep")
			badSpec("the hour has a decimal point", "3.2:09")
			badSpec("the minute has a decimal point", "03:0.9")
			badSpec("the hour has a space between the digits", "0 3:09")
			badSpec("the minute has a space between the digits", "03:0 9")
			badSpec("the timespec has extra characters", "03:09 PM beep")
			badSpec("there is an AM meridiem but the hour is 0", "0:09 AM")
			badSpec("there is a PM meridiem but the hour is 0", "0:09 PM")
		})

		Context("when the timespec is valid", func() {
			JustBeforeEach(func() {
				Expect(err).NotTo(HaveOccurred())
			})

			goodTests := func() {
				var afterTime = time.Date(2000, time.January, 1, 1, 0, 0, 0, time.UTC)
				//Always checks after 1AM on Jan 1st, 2000
				goodSpec := func(desc, spec string, hour, minute int) {
					Context(fmt.Sprintf("%s (%s)", desc, spec), func() {
						BeforeEach(func() {
							inputSpec = spec
						})

						It(fmt.Sprintf("should have the configured hour be %d", hour), func() {
							Expect(tod.Hour()).To(BeEquivalentTo(hour))
						})
						It(fmt.Sprintf("should have the configured minute be %d", minute), func() {
							Expect(tod.Minute()).To(BeEquivalentTo(minute))
						})

						Context(fmt.Sprintf("generating the next time after %s", afterTime.Format(time.RFC822)), func() {
							var nextTime time.Time
							JustBeforeEach(func() {
								nextTime = tod.NextAfter(afterTime)
							})

							It(fmt.Sprintf("should report that the next time's hour is %d", hour), func() {
								Expect(nextTime.Hour()).To(BeEquivalentTo(hour))
							})

							It(fmt.Sprintf("should report that the next time's minute is %d", minute), func() {
								Expect(nextTime.Minute()).To(BeEquivalentTo(minute))
							})

							Specify("the next time should be within the following day", func() {
								Expect(afterTime.Add(24*time.Hour + 1*time.Millisecond).Before(nextTime)).NotTo(BeTrue())
							})

							Specify("the next time should not be before the last time", func() {
								Expect(nextTime.Before(afterTime)).NotTo(BeTrue())
							})
						})
					})
				}

				Context("and there is no meridiem", func() {
					Context("and the time is 12th hour or before", func() {
						goodSpec("and there is no leading zero", "3:09", 3, 9)
						goodSpec("and there is a leading zero", "03:09", 3, 9)
						goodSpec("and the time is midnight", "0:00", 0, 0)
						goodSpec("and the time is midnight with a double-zero hour", "00:00", 0, 0)
						goodSpec("and there are leading spaces", " 3:09", 3, 9)
						goodSpec("and there are trailing spaces", "3:09 ", 3, 9)
						goodSpec("and there are leading and trailing spaces", " 3:09 ", 3, 9)
						goodSpec("and it's noon", "12:00", 12, 0)
						goodSpec("and the minute is 59", "3:59", 3, 59)
					})

					Context("and the time is after the 12th hour", func() {
						goodSpec("and the time is ", "18:13", 18, 13)
						goodSpec("and the time is midnight", "24:00", 0, 0)
						goodSpec("and the time is after midnight", "24:09", 0, 9)
					})
				})

				Context("and there is a meridiem", func() {
					Context("which is AM", func() {
						goodSpec("and there is a space between the minute and meridiem", "3:09 AM", 3, 9)
						goodSpec("and there is not a space between the minute and meridiem", "3:09AM", 3, 9)
						goodSpec("and the time is midnight", "12:00 AM", 0, 0)
						goodSpec("and the time is slightly after midnight", "12:01 AM", 0, 1)
						goodSpec("and there's a leading zero", "03:09 AM", 3, 9)
					})
					Context("which is PM", func() {
						goodSpec("and there is a space between the minute and meridiem", "3:09 PM", 15, 9)
						goodSpec("and there is not a space between the minute and meridiem", "3:09PM", 15, 9)
						goodSpec("and the time is noon", "12:00 PM", 12, 0)
						goodSpec("and the time is slightly after noon", "12:01 PM", 12, 1)
					})
				})
			}

			Context("when the location is UTC", func() {
				BeforeEach(func() {
					inputLoc = time.UTC
				})

				goodTests()
			})

			Context("when the location is not UTC", func() {
				BeforeEach(func() {
					inputLoc, err = time.LoadLocation("America/New_York")
					Expect(err).NotTo(HaveOccurred())
				})

				goodTests()
			})
		})
	})
})
