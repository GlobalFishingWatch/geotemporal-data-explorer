package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"net/http"
	"text/template"
	"time"

	"github.com/4wings/cli/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
)

var jsonBodyMonthGEE *template.Template
var jsonBodyDayGEE *template.Template

var INTERVALS_GEE = map[string]int{
	"hour":  0,
	"day":   1,
	"month": 2,
	"year":  3,
}

func init() {
	tmpl := template.New("jsonBodyMonthGEE")

	tmpl, err := tmpl.Parse(`{
  "expression": {
    "result": "0",
    "values": {
      "0": {
        "functionInvocationValue": {
          "arguments": {
            "collection": {
              "functionInvocationValue": {
                "arguments": {
                  "collection": {
                    "functionInvocationValue": {
                      "arguments": {
                        "collection": {
                          "functionInvocationValue": {
                            "arguments": {
                              "collection": {
                                "functionInvocationValue": {
                                  "arguments": {
                                    "collection": {
                                      "functionInvocationValue": {
                                        "arguments": {
                                          "id": {
                                            "constantValue": "{{.image}}"
                                          }
                                        },
                                        "functionName": "ImageCollection.load"
                                      }
                                    },
                                    "baseAlgorithm": {
                                      "functionDefinitionValue": {
                                        "argumentNames": ["_MAPPING_VAR_0_0"],
                                        "body": "1"
                                      }
                                    }
                                  },
                                  "functionName": "Collection.map"
                                }
                              },
                              "filter": {
                                "functionInvocationValue": {
                                  "arguments": {
                                    "rightField": { "valueReference": "2" },
                                    "leftValue": {
                                      "functionInvocationValue": {
                                        "arguments": {
                                          "start": {
                                            "constantValue": "{{.startDate}}"
                                          },
                                          "end": {
                                            "constantValue": "{{.endDate}}"
                                          }
                                        },
                                        "functionName": "DateRange"
                                      }
                                    }
                                  },
                                  "functionName": "Filter.dateRangeContains"
                                }
                              }
                            },
                            "functionName": "Collection.filter"
                          }
                        },
                        "key": { "valueReference": "2" },
                        "ascending": { "constantValue": true }
                      },
                      "functionName": "Collection.limit"
                    }
                  },
                  "baseAlgorithm": {
                    "functionDefinitionValue": {
                      "argumentNames": ["_MAPPING_VAR_0_0"],
                      "body": "3"
                    }
                  }
                },
                "functionName": "Collection.map"
              }
            },
            "baseAlgorithm": {
              "functionDefinitionValue": {
                "argumentNames": ["_MAPPING_VAR_0_0"],
                "body": "4"
              }
            }
          },
          "functionName": "Collection.map"
        }
      },
      "1": {
        "functionInvocationValue": {
          "arguments": {
            "input": { "argumentReference": "_MAPPING_VAR_0_0" },
            "bandSelectors": { "constantValue": ["{{.band}}"] }
          },
          "functionName": "Image.select"
        }
      },
      "2": { "constantValue": "system:time_start" },
      "3": {
        "functionInvocationValue": {
          "arguments": {
            "input": {
              "functionInvocationValue": {
                "arguments": {
                  "image": { "argumentReference": "_MAPPING_VAR_0_0" },
                  "crs": {
                    "functionInvocationValue": {
                      "arguments": { "crs": { "constantValue": "EPSG:4326" } },
                      "functionName": "Projection"
                    }
                  },
                  "crsTransform": { "constantValue": [1, 0, 0, 0, -1, 0] }
                },
                "functionName": "Image.setDefaultProjection"
              }
            },
            "geometry": {
              "functionInvocationValue": {
                "arguments": {
                  "coordinates": {
                    "constantValue": [
                      [
                        [{{.bounds.MinLon}}, {{.bounds.MaxLat}}],
                        [{{.bounds.MinLon}}, {{.bounds.MinLat}}],
                        [{{.bounds.MaxLon}}, {{.bounds.MinLat}}],
                        [{{.bounds.MaxLon}}, {{.bounds.MaxLat}}]
                      ]
                    ]
                  },
                  "geodesic": { "constantValue": false },
                  "evenOdd": { "constantValue": true }
                },
                "functionName": "GeometryConstructors.Polygon"
              }
            },
            "width": { "constantValue": {{.numCellsLon}} },
            "height": { "constantValue": {{.numCellsLat}} }
          },
          "functionName": "Image.clipToBoundsAndScale"
        }
      },
      "4": {
        "functionInvocationValue": {
          "arguments": {
            "image": { "argumentReference": "_MAPPING_VAR_0_0" },
            "min": { "constantValue": {{.min}} },
            "max": { "constantValue": {{.max}} }
          },
          "functionName": "Image.visualize"
        }
      }
    }
  },
  "fileFormat": "GIF",
  "videoOptions": { "framesPerSecond": 7 }
}
`)
	if err != nil {
		panic(err)
	}
	jsonBodyMonthGEE = tmpl

	tmpl, err = template.New("jsonBodyDayGEE").Parse(`{
  "expression": {
    "result": "0",
    "values": {
      "0": {
        "functionInvocationValue": {
          "arguments": {
            "collection": {
              "functionInvocationValue": {
                "arguments": {
                  "collection": {
                    "functionInvocationValue": {
                      "arguments": {
                        "images": {
                          "functionInvocationValue": {
                            "arguments": {
                              "list": {
                                "functionInvocationValue": {
                                  "arguments": {
                                    "start": { "valueReference": "1" },
                                    "end": { "constantValue": {{.limit}} }
                                  },
                                  "functionName": "List.sequence"
                                }
                              },
                              "baseAlgorithm": {
                                "functionDefinitionValue": {
                                  "argumentNames": ["_MAPPING_VAR_1_0"],
                                  "body": "2"
                                }
                              }
                            },
                            "functionName": "List.map"
                          }
                        }
                      },
                      "functionName": "ImageCollection.fromImages"
                    }
                  },
                  "baseAlgorithm": {
                    "functionDefinitionValue": {
                      "argumentNames": ["_MAPPING_VAR_0_0"],
                      "body": "11"
                    }
                  }
                },
                "functionName": "Collection.map"
              }
            },
            "baseAlgorithm": {
              "functionDefinitionValue": {
                "argumentNames": ["_MAPPING_VAR_0_0"],
                "body": "12"
              }
            }
          },
          "functionName": "Collection.map"
        }
      },
      "1": { "constantValue": 0 },
      "2": {
        "functionInvocationValue": {
          "arguments": {
            "condition": {
              "functionInvocationValue": {
                "arguments": {
                  "left": {
                    "functionInvocationValue": {
                      "arguments": { "collection": { "valueReference": "3" } },
                      "functionName": "Collection.size"
                    }
                  },
                  "right": { "valueReference": "1" }
                },
                "functionName": "Number.eq"
              }
            },
            "trueCase": {
              "functionInvocationValue": {
                "arguments": {
                  "object": {
                    "functionInvocationValue": {
                      "arguments": {
                        "input": {
                          "functionInvocationValue": {
                            "arguments": {
                              "image": { "valueReference": "8" },
                              "mask": { "valueReference": "8" }
                            },
                            "functionName": "Image.mask"
                          }
                        },
                        "names": { "valueReference": "9" }
                      },
                      "functionName": "Image.rename"
                    }
                  },
                  "key": { "valueReference": "4" },
                  "value": {
                    "functionInvocationValue": {
                      "arguments": { "date": { "valueReference": "6" } },
                      "functionName": "Date.millis"
                    }
                  }
                },
                "functionName": "Element.set"
              }
            },
            "falseCase": {
              "functionInvocationValue": {
                "arguments": {
                  "collection": {
                    "functionInvocationValue": {
                      "arguments": {
                        "collection": { "valueReference": "3" },
                        "baseAlgorithm": {
                          "functionDefinitionValue": {
                            "argumentNames": ["_MAPPING_VAR_0_0"],
                            "body": "10"
                          }
                        }
                      },
                      "functionName": "Collection.map"
                    }
                  }
                },
                "functionName": "reduce.mean"
              }
            }
          },
          "functionName": "If"
        }
      },
      "3": {
        "functionInvocationValue": {
          "arguments": {
            "collection": {
              "functionInvocationValue": {
                "arguments": {
                  "collection": {
                    "functionInvocationValue": {
                      "arguments": {
                        "id": {
                          "constantValue": "{{.image}}"
                        }
                      },
                      "functionName": "ImageCollection.load"
                    }
                  },
                  "filter": {
                    "functionInvocationValue": {
                      "arguments": {
                        "rightField": { "valueReference": "4" },
                        "leftValue": {
                          "functionInvocationValue": {
                            "arguments": {
                              "start": { "valueReference": "5" },
                              "end": { "constantValue": "{{.endDate}}" }
                            },
                            "functionName": "DateRange"
                          }
                        }
                      },
                      "functionName": "Filter.dateRangeContains"
                    }
                  }
                },
                "functionName": "Collection.filter"
              }
            },
            "filter": {
              "functionInvocationValue": {
                "arguments": {
                  "rightField": { "valueReference": "4" },
                  "leftValue": {
                    "functionInvocationValue": {
                      "arguments": {
                        "start": {
                          "functionInvocationValue": {
                            "arguments": {
                              "date": { "valueReference": "6" },
                              "delta": {
                                "argumentReference": "_MAPPING_VAR_1_0"
                              },
                              "unit": { "valueReference": "7" }
                            },
                            "functionName": "Date.advance"
                          }
                        },
                        "end": {
                          "functionInvocationValue": {
                            "arguments": {
                              "date": { "valueReference": "6" },
                              "delta": {
                                "functionInvocationValue": {
                                  "arguments": {
                                    "left": {
                                      "argumentReference": "_MAPPING_VAR_1_0"
                                    },
                                    "right": { "constantValue": 1 }
                                  },
                                  "functionName": "Number.add"
                                }
                              },
                              "unit": { "valueReference": "7" }
                            },
                            "functionName": "Date.advance"
                          }
                        }
                      },
                      "functionName": "DateRange"
                    }
                  }
                },
                "functionName": "Filter.dateRangeContains"
              }
            }
          },
          "functionName": "Collection.filter"
        }
      },
      "4": { "constantValue": "system:time_start" },
      "5": { "constantValue": "{{.startDate}}" },
      "6": {
        "functionInvocationValue": {
          "arguments": { "value": { "valueReference": "5" } },
          "functionName": "Date"
        }
      },
      "7": { "constantValue": "day" },
      "8": {
        "functionInvocationValue": {
          "arguments": { "value": { "constantValue": 0 } },
          "functionName": "Image.constant"
        }
      },
      "9": { "constantValue": ["{{.band}}"] },
      "10": {
        "functionInvocationValue": {
          "arguments": {
            "input": { "argumentReference": "_MAPPING_VAR_0_0" },
            "bandSelectors": { "valueReference": "9" }
          },
          "functionName": "Image.select"
        }
      },
      "11": {
        "functionInvocationValue": {
          "arguments": {
            "input": {
              "functionInvocationValue": {
                "arguments": {
                  "image": { "argumentReference": "_MAPPING_VAR_0_0" },
                  "crs": {
                    "functionInvocationValue": {
                      "arguments": { "crs": { "constantValue": "EPSG:4326" } },
                      "functionName": "Projection"
                    }
                  },
                  "crsTransform": { "constantValue": [1, 0, 0, 0, -1, 0] }
                },
                "functionName": "Image.setDefaultProjection"
              }
            },
            "geometry": {
              "functionInvocationValue": {
                "arguments": {
                  "coordinates": {
                    "constantValue": [
                      [
                        [{{.bounds.MinLon}}, {{.bounds.MaxLat}}],
                        [{{.bounds.MinLon}}, {{.bounds.MinLat}}],
                        [{{.bounds.MaxLon}}, {{.bounds.MinLat}}],
                        [{{.bounds.MaxLon}}, {{.bounds.MaxLat}}]
                      ]
                    ]
                  },
                  "geodesic": { "constantValue": false },
                  "evenOdd": { "constantValue": true }
                },
                "functionName": "GeometryConstructors.Polygon"
              }
            },
            "width": { "constantValue": {{.numCellsLon}} },
            "height": { "constantValue": {{.numCellsLat}} }
          },
          "functionName": "Image.clipToBoundsAndScale"
        }
      },
      "12": {
        "functionInvocationValue": {
          "arguments": {
            "image": { "argumentReference": "_MAPPING_VAR_0_0" },
            "min": { "constantValue": {{.min}} },
            "max": { "constantValue": {{.max}} }
          },
          "functionName": "Image.visualize"
        }
      }
    }
  },
  "fileFormat": "GIF",
  "videoOptions": { "framesPerSecond": 7 }
}
`)
	if err != nil {
		panic(err)
	}
	jsonBodyDayGEE = tmpl
}

func generateGIF(z, x, y, numCellsLat, numCellsLon int, dataset *types.Dataset, startDate, endDate time.Time, interval string, limit int) (*gif.GIF, error) {
	conf, err := google.JWTConfigFromJSON([]byte(viper.GetString("gee-account")))
	if err != nil {
		return nil, err
	}
	conf.Scopes = []string{
		"https://www.googleapis.com/auth/earthengine",
	}
	log.Debugf("Using image %s and band %s between %.2f, %.2f", dataset.Configuration.Images[INTERVALS_GEE[interval]], dataset.Configuration.Band, dataset.Configuration.Min, dataset.Configuration.Max)
	client := conf.Client(context.TODO())
	var body bytes.Buffer
	paramsQuery := map[string]interface{}{
		"min":         dataset.Configuration.Min,
		"max":         dataset.Configuration.Max,
		"numCellsLat": numCellsLat,
		"numCellsLon": numCellsLon,
		"band":        dataset.Configuration.Band,
		"startDate":   startDate.Format("2006-01-02"),
		"endDate":     endDate.Format("2006-01-02"),
		"image":       dataset.Configuration.Images[INTERVALS_GEE[interval]],
		"bounds":      TileToBBOX(x, y, z),
		"limit":       limit - 1,
	}
	if interval == "day" {
		err = jsonBodyDayGEE.Execute(&body, paramsQuery)
	} else if interval == "month" {
		err = jsonBodyMonthGEE.Execute(&body, paramsQuery)
	} else {
		return nil, fmt.Errorf("interval %s not supported", interval)
	}
	if err != nil {
		return nil, err
	}
	resp, err := client.Post("https://earthengine.googleapis.com/v1alpha/projects/earthengine-legacy/videoThumbnails?fields=name", "application/json", bytes.NewBuffer(body.Bytes()))
	if err != nil {
		return nil, err
	}
	var jsonBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&jsonBody)
	if err != nil {
		return nil, err
	}
	if jsonBody["error"] != nil {
		log.Error("Error obtaining image id", jsonBody["error"].(map[string]interface{})["message"])
		return nil, fmt.Errorf(fmt.Sprintf("%s", jsonBody["error"].(map[string]interface{})["message"]))
	}
	log.Debugf("obtaining gif: %s", fmt.Sprintf("https://earthengine.googleapis.com/v1alpha/%s:getPixels", jsonBody["name"]))
	resp, err = http.Get(fmt.Sprintf("https://earthengine.googleapis.com/v1alpha/%s:getPixels", jsonBody["name"]))
	if err != nil {
		log.Error("Error obtaining gif", err)
		return nil, err
	}

	defer resp.Body.Close()

	return gif.DecodeAll(resp.Body)

}

func getGifDimensions(gif *gif.GIF) (x, y int) {
	var lowestX int
	var lowestY int
	var highestX int
	var highestY int

	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}

	return highestX - lowestX, highestY - lowestY
}

func ReadGEE(dataset *types.Dataset, z, x, y int, temporalAggregation bool, dateRange string, interval string) ([][]int, error) {

	pos, err := Tile2Num(z, x, y)
	if err != nil {
		log.Error("Error obtaining pos", err)
		return nil, err
	}
	numCellsLat, numCellsLon := GetCellsLatLonByPos(int(pos), z)
	startDate, endDate, err := ParseDateRange(dateRange)
	if err != nil {
		return nil, err
	}
	var limit int
	if interval == "month" {
		limit = MonthDiff(startDate, endDate)
	} else if interval == "day" {
		limit = DaysDiff(startDate, endDate)
	}
	var g *gif.GIF
	for i := 0; i < 3; i++ {
		g, err = generateGIF(z, x, y, numCellsLat, numCellsLon, dataset, startDate, endDate, interval, limit)
		if err != nil {
			log.Error("Error generating gif", err)
			time.Sleep(1 * time.Second)
			continue
		} else {
			break
		}
	}
	if err != nil {
		log.Error("Error generating gif", err)
		return nil, err
	}
	var list [][]int

	numCells := numCellsLat * numCellsLon
	if temporalAggregation {
		list = make([][]int, numCells)
	} else {
		list = make([][]int, numCells*len(g.Image))
	}
	total := dataset.Configuration.Max - dataset.Configuration.Min
	step := float64(total) / float64(65536)

	if len(g.Image) != limit {
		return nil, fmt.Errorf("num layers different of dates. Sent %d Received %d", limit, len(g.Image))
	}
	imgWidth, imgHeight := getGifDimensions(g)

	overpaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	fmt.Println(" g.Image", len(g.Image))
	for numImage, img := range g.Image {
		draw.Draw(overpaintImage, overpaintImage.Bounds(), img, image.ZP, draw.Over)

		for i := 0; i < numCellsLat; i++ {
			for j := 0; j < numCellsLon; j++ {
				index := 0
				if temporalAggregation {
					index = (i * numCellsLon) + j
				} else {
					index = ((i * numCellsLon) + j) + (numImage * numCells)
				}
				if len(list[index]) == 0 {
					list[index] = []int{0, 0, 0}
				}

				r, _, _, _ := overpaintImage.At(j, numCellsLat-(i+1)).RGBA()
				if r > 0 {
					value := ((float64(r)*step + float64(dataset.Configuration.Min)) * dataset.Configuration.Scale) + dataset.Configuration.Offset
					if value > 0 {
						list[index][0] += int(value * float64(100))
						list[index][1]++
						list[index][2] = 1
					}
				}
			}
		}
	}
	return list, nil

}
