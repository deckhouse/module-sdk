package hookinfolder_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "example-module/subfolder"
)

var _ = Describe("snapshot example", func() {
	Context("refoncile func", func() {
		When("all services works correctly", func() {
			snapshots := mock.NewSnapshotsMock(GinkgoT())
			snapshots.GetMock.When(subfolder.NodeInfoSnapshotName).Then(
				[]pkg.Snapshot{
					mock.NewSnapshotMock(GinkgoT()).UnmarshalToMock.Set(func(v any) error {
						node := v.(*subfolder.NodeInfo)
						*node = subfolder.NodeInfo{
							APIVersion: "v1",
							Kind:       "node",
							Metadata: subfolder.NodeInfoMetadata{
								Name:            "first-node",
								ResourceVersion: "v1",
								UID:             "1",
							},
						}

						return nil
					}),
					mock.NewSnapshotMock(GinkgoT()).UnmarshalToMock.Set(func(v any) error {
						node := v.(*subfolder.NodeInfo)
						*node = subfolder.NodeInfo{
							APIVersion: "v1",
							Kind:       "node",
							Metadata: subfolder.NodeInfoMetadata{
								Name:            "second-node",
								ResourceVersion: "v1",
								UID:             "2",
							},
						}

						return nil
					}),
				},
			)

			buf := bytes.NewBuffer([]byte{})

			var input = &pkg.HookInput{
				Snapshots: snapshots,
				Logger: log.NewLogger(log.Options{
					Level:  log.LevelDebug.Level(),
					Output: buf,
					TimeFunc: func(_ time.Time) time.Time {
						parsedTime, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
						Expect(err).ShouldNot(HaveOccurred())

						return parsedTime
					},
				}),
			}

			It("reconcile func executed correctly", func() {
				err := subfolder.HandlerHookSnapshots(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())

				logs := strings.Split(buf.String(), "\n")

				Expect(logs[0]).Should(ContainSubstring(`"level":"info","msg":"hello from snapshot hook"`))
				Expect(logs[1]).Should(ContainSubstring(`"level":"info","msg":"node found"`))
				Expect(logs[1]).Should(ContainSubstring(`"APIVersion":"v1","Kind":"node","Name":"first-node","ResourceVersion":"v1","UID":"1"`))
				Expect(logs[2]).Should(ContainSubstring(`"level":"info","msg":"node found"`))
				Expect(logs[2]).Should(ContainSubstring(`"APIVersion":"v1","Kind":"node","Name":"second-node","ResourceVersion":"v1","UID":"2"`))
			})
		})

		When("unmarshal get error", func() {
			snapshots := mock.NewSnapshotsMock(GinkgoT())
			snapshots.GetMock.When(subfolder.NodeInfoSnapshotName).Then(
				[]pkg.Snapshot{
					mock.NewSnapshotMock(GinkgoT()).UnmarshalToMock.Set(func(_ any) error {
						return errors.New("error")
					}),
				},
			)

			buf := bytes.NewBuffer([]byte{})

			var input = &pkg.HookInput{
				Snapshots: snapshots,
				Logger: log.NewLogger(log.Options{
					Level:  log.LevelDebug.Level(),
					Output: buf,
					TimeFunc: func(_ time.Time) time.Time {
						parsedTime, err := time.Parse(time.DateTime, "2006-01-02 15:04:05")
						Expect(err).ShouldNot(HaveOccurred())

						return parsedTime
					},
				}),
			}

			It("unmarshal returns error", func() {
				err := subfolder.HandlerHookSnapshots(context.Background(), input)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(Equal(fmt.Errorf("unmarshal to: %w", errors.New("error"))))

				logs := strings.Split(buf.String(), "\n")

				Expect(logs[0]).Should(ContainSubstring(`"level":"info","msg":"hello from snapshot hook"`))
			})
		})
	})
})
