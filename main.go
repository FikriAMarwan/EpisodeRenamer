package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/facette/natsort"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type MyMainWindow struct {
	*walk.MainWindow
}

type Episode struct {
	Index   int
	Episode string
	Name    string
	OVA     string
	OVABool bool
}

type EpisodeModel struct {
	walk.SortedReflectTableModelBase
	items []*Episode
}

func (m *EpisodeModel) Items() interface{} {
	return m.items
}

func NewEpisodeModel() *EpisodeModel {
	m := new(EpisodeModel)
	return m
}

func main() {
	var anime, target *walk.TextEdit
	var tv *walk.TableView
	model := new(EpisodeModel)
	mw := new(MyMainWindow)
	MainWindow{
		Title:    "Episode Renamer By Grisaria Ver 0.1",
		AssignTo: &mw.MainWindow,
		Size:     Size{500, 600},
		Layout:   Grid{},
		OnDropFiles: func(folder []string) {
			if len(folder) == 1 {
				files, err := FindAnime(folder[0])
				if err != nil {
					if err.Error() == "File" {
						walk.MsgBox(mw, "Episode Renamer", "Please Drops Folder Only", walk.MsgBoxIconWarning)
					}
					walk.MsgBox(mw, "Episode Renamer", "Something Went Wrong", walk.MsgBoxIconWarning)
					return
				}
				if len(files) == 0 {
					walk.MsgBox(mw, "Episode Renamer", "No Video Found", walk.MsgBoxIconWarning)
				}
				anime.SetText(folder[0])
				model.Add(files)
			} else {
				walk.MsgBox(mw, "Episode Renamer", "MAX Drop 1", walk.MsgBoxIconWarning)
			}
		},
		Children: []Widget{
			HSplitter{
				MaxSize: Size{500, 20},
				Row:     1,
				Children: []Widget{
					PushButton{
						MaxSize: Size{20, 20},
						Text:    "Select Anime Folder",
						OnClicked: func() {
							dlg := new(walk.FileDialog)
							dlg.Title = "Anime Folder"
							if ok, err := dlg.ShowBrowseFolder(mw); err != nil {
								walk.MsgBox(mw, "Episode Renamer", "Something Went Wrong", walk.MsgBoxIconWarning)
							} else if !ok {
								anime.SetText("No Folder Selected")
							} else {
								files, err := FindAnime(dlg.FilePath)
								if err != nil {
									walk.MsgBox(mw, "Episode Renamer", "Something Went Wrong", walk.MsgBoxIconWarning)
								}
								if len(files) == 0 {
									walk.MsgBox(mw, "Episode Renamer", "No Video Found", walk.MsgBoxIconWarning)
								}
								model.Add(files)
								anime.SetText(dlg.FilePath)
							}
						},
					},
					TextEdit{MinSize: Size{50, 20}, AssignTo: &anime, ReadOnly: true, Text: "Select Folder"},
				},
			},
			HSplitter{
				MaxSize: Size{500, 20},
				Row:     2,
				Children: []Widget{
					Label{Text: "Nama Anime : "},
					TextEdit{AssignTo: &target},
				},
			},
			TableView{
				AssignTo:                 &tv,
				Row:                      3,
				MinSize:                  Size{500, 500},
				Model:                    model,
				AlternatingRowBG:         true,
				CheckBoxes:               true,
				AlwaysConsumeSpace:       true,
				NotSortableByHeaderClick: true,
				// MultiSelection: true,
				Columns: []TableViewColumn{
					{Name: "Index", Width: 60, Hidden: true},
					{Name: "Episode", Width: 60},
					{Name: "Name", Width: 387},
					{Name: "OVA", Width: 50},
				},
				OnKeyPress: func(key walk.Key) {
					i := tv.SelectedIndexes()
					if key == walk.KeyDelete || key == walk.KeyBack {
						for _, v := range i {
							model.Delete(v)
						}
						model.Refresh()
						model.PublishRowsChanged(0, len(model.items))
					}
					if key == walk.KeyUp && len(i) < len(model.items) {
						for _, v := range i {
							model.Move(v, "UP")
						}
						tv.SetSelectedIndexes(i)
					}
					if key == walk.KeyDown && len(i) < len(model.items) {
						sort.Sort(sort.Reverse(sort.IntSlice(i)))
						for _, v := range i {
							model.Move(v, "DOWN")
						}
						tv.SetSelectedIndexes(i)
					}
					tv.SetSelectedIndexes(i)
				},
			},
			PushButton{
				Text: "RENAME IT",
				Row:  4,
				OnClicked: func() {
					if anime.Text() == "No Folder Selected" || anime.Text() == "Select Folder" {
						walk.MsgBox(mw, "Episode Renamer", "Select The Folder", walk.MsgBoxIconWarning)
						return
					}
					if len(target.Text()) == 0 {
						walk.MsgBox(mw, "Episode Renamer", "Insert The Title", walk.MsgBoxIconWarning)
						return
					}
					err := Rename(model.items, anime.Text(), target.Text())
					if err != nil {
						if err.Error() == "Folder Not Found" {
							walk.MsgBox(mw, "Episode Renamer", "Folder Not Found, Please Select Again", walk.MsgBoxIconWarning)
							return
						}
						walk.MsgBox(mw, "Episode Renamer", "Something Went Wrong", walk.MsgBoxIconWarning)
						return
					}
					walk.MsgBox(mw, "Episode Renamer", "Episode Has Been Renamed To "+target.Text(), walk.MsgBoxOK)
					anime.SetText("No Folder Selected")
					target.SetText("")
					model.items = []*Episode{}
					model.PublishRowsReset()
				},
			},
		},
	}.Run()

}

func (m *EpisodeModel) Checked(row int) bool {
	return m.items[row].OVABool
}

func (m *EpisodeModel) SetChecked(row int, checked bool) error {
	m.items[row].OVABool = checked
	m.Refresh()
	return nil
}

func (m *EpisodeModel) Refresh() {
	EpOva := 1
	Ep := 1
	for _, v := range m.items {
		if v.OVABool {
			v.Index = EpOva
			v.Episode = fmt.Sprintf("OVA %d", EpOva)
			EpOva++
		} else {
			v.Index = len(m.items) + Ep
			v.Episode = fmt.Sprintf("%d", Ep)
			Ep++
		}
	}
	m.PublishRowsChanged(0, len(m.items))
}

func (m *EpisodeModel) Add(files []string) {
	m.items = make([]*Episode, len(files))
	for i := range m.items {
		m.items[i] = &Episode{
			Index:   len(files) + i,
			Episode: fmt.Sprintf("%v", i+1),
			Name:    files[i],
			OVA:     "Not OVA",
			OVABool: false,
		}
	}
	m.PublishRowsReset()
}

func (m *EpisodeModel) Delete(i int) {
	if i >= 0 {
		m.items = append(m.items[:i], m.items[i+1:]...)
		m.PublishRowsRemoved(i, i)
	}
}

func (m *EpisodeModel) Move(i int, way string) {
	if i > 0 && i <= len(m.items)-1 && way == "UP" {
		temp := m.items[i-1].Name
		m.items[i-1].Name = m.items[i].Name
		m.items[i].Name = temp
	}
	if i >= 0 && i < len(m.items)-1 && way == "DOWN" {
		temp := m.items[i+1].Name
		m.items[i+1].Name = m.items[i].Name
		m.items[i].Name = temp
	}
	m.PublishRowsChanged(0, len(m.items))
}

func FindAnime(path string) (files []string, err error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("File")
	}
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".mkv" || strings.ToLower(filepath.Ext(path)) == ".mp4" {
			files = append(files, info.Name())
		}
		return nil
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	natsort.Sort(files)
	return files, nil
}

func Rename(list []*Episode, path string, target string) error {
	var final string
	if target[len(target)-1:] != " " {
		target = target + " "
	}

	if path[len(path)-1:] != `\` {
		path = path + `\`
	}
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Folder Not Found")
		}
	}
	for _, file := range list {
		ext := filepath.Ext(file.Name)
		origin := path + file.Name
		if file.OVABool {
			final = path + target + "Episode " + file.Episode + ext
		} else {
			final = path + target + "Episode " + file.Episode + ext
		}
		err := os.Rename(origin, final)
		if err != nil {
			return err
		}
	}

	return nil
}
