package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	//cameras, _ := getCameraMap()
	//fmt.Printf("%+v\n", cameras)

	s := server.NewMCPServer(
		"Reminder Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)
	// Add tool
	tool := mcp.NewTool("createReminder",
		mcp.WithDescription("creates a new to do reminder for a user"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("the title for the reminder, if the user did not provide a title create a short tile based on the reminder content"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("the contents of the reminder"),
		),
	)
	s.AddTool(tool, createReminder)

	tool = mcp.NewTool("listReminders",
		//mcp.WithDescription("captures a live image from a camera"),
		mcp.WithDescription("returns a list of reminders"),
	)
	s.AddTool(tool, listReminders)

	tool = mcp.NewTool("deleteReminder",
		//mcp.WithDescription("captures a live image from a camera"),
		mcp.WithDescription("deletes a reminder"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("the title for the reminder to delete"),
		),
	)
	s.AddTool(tool, deleteReminder)

	s.AddTool(mcp.NewTool("showReminder",
		//mcp.WithDescription("captures a live image from a camera"),
		mcp.WithDescription("show the contents of the reminder"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("the title for the reminder"),
		),
	), showReminder)

	s.AddTool(
		mcp.NewTool("completeReminders",
			//mcp.WithDescription("captures a live image from a camera"),
			mcp.WithDescription("completes a reminder"),
			mcp.WithString("title",
				mcp.Required(),
				mcp.Description("the title for the reminder to complete"),
			),
		), completeReminder)

	//tool =
	s.AddTool(
		mcp.NewTool("listCompleteReminders",
			mcp.WithDescription("lists recently completed reminders"),
		),
		listCompletedReminders)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

type Reminder struct {
	Title     string
	Date      time.Time
	Completed time.Time
	Content   string
}

type ReminderList struct {
	Reminders []Reminder
	Completed []Reminder
}

var reminderFile = "reminders.json"

func loadReminders() *ReminderList {
	reminders := &ReminderList{}

	data, err := os.ReadFile(reminderFile)
	if err != nil {
		return reminders
	}

	err = json.Unmarshal(data, reminders)
	if err != nil {
		return reminders
	}

	return reminders
}

func saveReminders(reminders *ReminderList) error {
	data, err := json.MarshalIndent(reminders, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(reminderFile, data, 0644)
}

func createReminder(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, ok := request.Params.Arguments["title"].(string)
	if !ok {
		return nil, errors.New("the title must be a string")
	}

	content, ok := request.Params.Arguments["content"].(string)
	if !ok {
		return nil, errors.New("the content needs to be a string")
	}

	reminderList := loadReminders()

	reminderList.Reminders = append(reminderList.Reminders, Reminder{
		Title:   title,
		Content: content,
		Date:    time.Now(),
	})

	saveReminders(reminderList)

	return mcp.NewToolResultText(fmt.Sprintf("a reminder titled %s was created", title)), nil
}

func deleteReminder(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, ok := request.Params.Arguments["title"].(string)
	if !ok {
		return nil, errors.New("the title must be a string")
	}

	reminderList := loadReminders()

	found := false
	newReminderList := &ReminderList{}
	for _, reminder := range reminderList.Reminders {
		if strings.ToLower(reminder.Title) == strings.ToLower(title) {
			found = true
			continue
		}
		newReminderList.Reminders = append(newReminderList.Reminders, reminder)
	}

	if !found {
		return mcp.NewToolResultText(fmt.Sprintf("i could not find the reminder titled %s", title)), nil
	}

	if err := saveReminders(newReminderList); err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("the reminder titled %s could not be deleted", title)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("the reminder titled %s was deleted", title)), nil
}

func listReminders(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	reminderList := loadReminders()

	if len(reminderList.Reminders) == 0 {
		return mcp.NewToolResultText("you have no reminders"), nil
	}

	activeList := "The user has the following reminders:"
	for _, reminder := range reminderList.Reminders {
		activeList += fmt.Sprintf("\n* %s", reminder.Title)
	}

	return mcp.NewToolResultText(activeList), nil
}

func showReminder(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, ok := request.Params.Arguments["title"].(string)
	if !ok {
		return nil, errors.New("the title must be a string")
	}

	reminderList := loadReminders()

	for _, reminder := range reminderList.Reminders {
		if strings.ToLower(reminder.Title) == strings.ToLower(title) {
			return mcp.NewToolResultText(reminder.Content), nil
		}
	}

	return mcp.NewToolResultText(fmt.Sprintf("i could not find the reminder titled %s", title)), nil
}

func completeReminder(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, ok := request.Params.Arguments["title"].(string)
	if !ok {
		return nil, errors.New("the title must be a string")
	}

	reminderList := loadReminders()

	found := false
	newReminderList := &ReminderList{}
	for idx := range reminderList.Reminders {
		reminder := reminderList.Reminders[idx]
		if strings.ToLower(reminder.Title) == strings.ToLower(title) {
			found = true
			reminder.Completed = time.Now()
			newReminderList.Completed = append(newReminderList.Completed, reminder)
			continue
		}
		newReminderList.Reminders = append(newReminderList.Reminders, reminder)
	}

	if !found {
		return mcp.NewToolResultText(fmt.Sprintf("i could not find the reminder titled %s", title)), nil
	}

	if err := saveReminders(newReminderList); err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("the reminder titled %s could not be completed", title)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("the reminder titled %s was completed", title)), nil
}

func listCompletedReminders(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	reminderList := loadReminders()

	if len(reminderList.Completed) == 0 {
		return mcp.NewToolResultText("you have no completed reminders yet"), nil
	}

	completedList := "The user has the completed the following reminders:"
	for _, reminder := range reminderList.Completed {
		completedList += fmt.Sprintf("\n%s completed on %s", reminder.Title, reminder.Completed.Format("January 2, 2006"))
	}

	return mcp.NewToolResultText(completedList), nil
}
