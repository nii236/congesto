package main

import (
	"reflect"
	"testing"
)

func Test_statusDiff(t *testing.T) {
	type args struct {
		prev *Status
		next *Status
	}
	tests := []struct {
		name    string
		args    args
		want    []*UpdatedServer
		wantErr bool
	}{
		{
			name: "Changed Category",
			want: []*UpdatedServer{&UpdatedServer{
				Region:     "Japan",
				Server:     "Tonberry",
				DataCentre: "Elemental",
				Key:        "Category",
				From:       "Congested",
				To:         "Preferred",
			}},
			wantErr: false,
			args: args{
				prev: &Status{
					Regions: []*Region{
						&Region{
							Name: "Japan",
							DataCentres: []*DataCentre{
								&DataCentre{
									Name: "Elemental",
									Servers: []*Server{
										&Server{
											Name:                     "Tonberry",
											Category:                 "Congested",
											CreateCharacterAvailable: true,
										},
									},
								},
							},
						},
					},
				},
				next: &Status{
					Regions: []*Region{
						&Region{
							Name: "Japan",
							DataCentres: []*DataCentre{
								&DataCentre{
									Name: "Elemental",
									Servers: []*Server{
										&Server{
											Name:                     "Tonberry",
											Category:                 "Preferred",
											CreateCharacterAvailable: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "Changed CreateCharacterAvailable",
			wantErr: false,
			want: []*UpdatedServer{&UpdatedServer{
				Region:     "Japan",
				Server:     "Tonberry",
				DataCentre: "Elemental",
				Key:        "CreateCharacterAvailable",
				From:       false,
				To:         true,
			}},
			args: args{
				prev: &Status{
					Regions: []*Region{
						&Region{
							Name: "Japan",
							DataCentres: []*DataCentre{
								&DataCentre{
									Name: "Elemental",
									Servers: []*Server{
										&Server{
											Name:                     "Tonberry",
											Category:                 "Congested",
											CreateCharacterAvailable: false,
										},
									},
								},
							},
						},
					},
				},
				next: &Status{
					Regions: []*Region{
						&Region{
							Name: "Japan",
							DataCentres: []*DataCentre{
								&DataCentre{
									Name: "Elemental",
									Servers: []*Server{
										&Server{
											Name:                     "Tonberry",
											Category:                 "Congested",
											CreateCharacterAvailable: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "Changed both",
			wantErr: false,
			want: []*UpdatedServer{
				&UpdatedServer{
					Region:     "Japan",
					Server:     "Tonberry",
					DataCentre: "Elemental",
					Key:        "Category",
					From:       "Congested",
					To:         "Preferred",
				},
				&UpdatedServer{
					Region:     "Japan",
					Server:     "Tonberry",
					DataCentre: "Elemental",
					Key:        "CreateCharacterAvailable",
					From:       false,
					To:         true,
				},
			},
			args: args{
				prev: &Status{
					Regions: []*Region{
						&Region{
							Name: "Japan",
							DataCentres: []*DataCentre{
								&DataCentre{
									Name: "Elemental",
									Servers: []*Server{
										&Server{
											Name:                     "Tonberry",
											Category:                 "Congested",
											CreateCharacterAvailable: false,
										},
									},
								},
							},
						},
					},
				},
				next: &Status{
					Regions: []*Region{
						&Region{
							Name: "Japan",
							DataCentres: []*DataCentre{
								&DataCentre{
									Name: "Elemental",
									Servers: []*Server{
										&Server{
											Name:                     "Tonberry",
											Category:                 "Preferred",
											CreateCharacterAvailable: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:    "No change",
			wantErr: false,
			want:    nil,
			args: args{
				prev: &Status{
					Regions: []*Region{
						&Region{
							Name: "Japan",
							DataCentres: []*DataCentre{
								&DataCentre{
									Name: "Elemental",
									Servers: []*Server{
										&Server{
											Name:                     "Tonberry",
											Category:                 "Congested",
											CreateCharacterAvailable: true,
										},
									},
								},
							},
						},
					},
				},
				next: &Status{
					Regions: []*Region{
						&Region{
							Name: "Japan",
							DataCentres: []*DataCentre{
								&DataCentre{
									Name: "Elemental",
									Servers: []*Server{
										&Server{
											Name:                     "Tonberry",
											Category:                 "Congested",
											CreateCharacterAvailable: true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := statusDiff(tt.args.prev, tt.args.next)
			if err != nil {
				t.Errorf("statusDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, v := range got {
				if !reflect.DeepEqual(v, tt.want[i]) {
					t.Errorf("statusDiff() = %v, want %v", v, tt.want[i])
				}
			}

		})
	}
}
