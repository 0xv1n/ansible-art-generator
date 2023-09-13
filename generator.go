package main

import (
	"fmt"
	"strings"
	"gopkg.in/yaml.v3"
	fyne "fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"os"
)

type Apply struct {
	Become string `yaml:"become,omitempty"`
}

type AnsibleTask struct {
	IncludeRole struct {
		Name  string `yaml:"name"`
		Apply Apply  `yaml:"apply"`
	} `yaml:"include_role"`
	Vars map[string][]string `yaml:"vars"`
}

type AnsiblePlaybook struct {
	Hosts       string        `yaml:"hosts"`
	GatherFacts bool          `yaml:"gather_facts"`
	Tasks       []AnsibleTask `yaml:"tasks"`
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("ART: Ansible Playbook Generator")
	myWindow.Resize(fyne.NewSize(600, 400))

	var stringEntries []string
	var artWindow fyne.Window
	artTable := container.NewGridWithColumns(4)

	var updateStringTable func()
	updateStringTable = func() {
		artTable.Objects = nil
		for i, entry := range stringEntries {
			index := i
			label := widget.NewLabel(entry)
			upButton := widget.NewButton("Up", func() {
				if index > 0 {
					stringEntries[index-1], stringEntries[index] = stringEntries[index], stringEntries[index-1]
					updateStringTable()
				}
			})
			downButton := widget.NewButton("Down", func() {
				if index < len(stringEntries)-1 {
					stringEntries[index], stringEntries[index+1] = stringEntries[index+1], stringEntries[index]
					updateStringTable()
				}
			})
			deleteButton := widget.NewButton("Delete", func() {
				stringEntries = append(stringEntries[:index], stringEntries[index+1:]...)
				updateStringTable()
			})
			artTable.Add(label)
			artTable.Add(upButton)
			artTable.Add(downButton)
			artTable.Add(deleteButton)
		}
		artTable.Refresh()
	}

	filenameEntry := widget.NewEntry()
	filenameEntry.SetPlaceHolder("output.yml")

	osCombo := widget.NewSelect([]string{"linux", "windows", "all"}, nil)
	hostCombo := widget.NewSelect([]string{"all", "workstations", "servers"}, nil)
	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Enter playbook description here...")

	artCheckbox := widget.NewCheck("ART Test", func(checked bool) {
		if checked {
			artWindow = myApp.NewWindow("ART TIDs")
			artWindow.Resize(fyne.NewSize(600, 400))
			artEntry := widget.NewEntry()
			artEntry.SetPlaceHolder("Enter ART TID Here:")
			addButton := widget.NewButton("Add", func() {
				stringEntries = append(stringEntries, artEntry.Text)
				artEntry.SetText("")
				updateStringTable()
			})
			artWindow.SetContent(container.NewVBox(
				artEntry,
				addButton,
				artTable,
			))
			artWindow.Show()
		} else {
			if artWindow != nil {
				artWindow.Close()
			}
		}
	})

	generateButton := widget.NewButton("Generate YAML", func() {
		filename := filenameEntry.Text
		selectedOs := osCombo.Selected
		selectedHosts := hostCombo.Selected
		description := descriptionEntry.Text

		task := AnsibleTask{}
		task.IncludeRole.Name = "art-execution-role"
		task.IncludeRole.Apply.Become = "{{ 'True' if ansible_facts['system'] == 'Linux' else 'False' }}"
		task.Vars = make(map[string][]string)
		task.Vars["art_tids_"+selectedOs] = stringEntries

		playbook := AnsiblePlaybook{
			Hosts:       selectedHosts,
			GatherFacts: true,
			Tasks:       []AnsibleTask{task},
		}

		yamlData, err := yaml.Marshal(&playbook)
		if err != nil {
			fmt.Println(err)
			return
		}

		file, err := os.Create(filename)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		file.WriteString("---\n")
		if description != "" {
			formattedDescription := "# " + strings.ReplaceAll(description, "\n", "\n# ")
			file.WriteString(formattedDescription + "\n")
		}
		file.Write(yamlData)
	})

	myWindow.SetContent(
		container.NewVBox(
			filenameEntry,
			osCombo,
			hostCombo,
			descriptionEntry,
			artCheckbox,
			generateButton,
		),
	)
	myWindow.ShowAndRun()
}
