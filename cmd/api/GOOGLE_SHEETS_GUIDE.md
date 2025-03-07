# Google Sheets Email Verification Guide

This guide provides detailed instructions for setting up and using the Email Verifier Google Sheets integration.

## Table of Contents

1. [Overview](#overview)
2. [Setup Instructions](#setup-instructions)
3. [Using the Add-on](#using-the-add-on)
4. [Custom Functions](#custom-functions)
5. [Example Workflows](#example-workflows)
6. [Troubleshooting](#troubleshooting)
7. [Advanced Usage](#advanced-usage)

## Overview

The Email Verifier Google Sheets integration allows you to verify email addresses directly within your Google Sheets without having to download, process, and re-upload your data. This integration uses a custom Google Apps Script that communicates with your deployed Email Verifier API.

Key features:
- Verify selected email cells
- Verify entire columns of emails
- Automatically detect and verify emails in a named column
- Use custom functions to verify emails directly in formulas
- Configure the API endpoint
- Display verification results directly in your spreadsheet

## Setup Instructions

### Step 1: Deploy the Email Verifier API

Before setting up the Google Sheets integration, you need to deploy the Email Verifier API. Follow the instructions in the [API Documentation](API_DOCUMENTATION.md) to deploy the API to a cloud provider like Render, Google Cloud Run, or AWS.

Make note of the URL where your API is deployed (e.g., `https://email-verifier-api.onrender.com`).

### Step 2: Set Up the Google Apps Script

1. Open your Google Sheet containing email addresses
2. Click on **Extensions** > **Apps Script** in the menu bar
3. In the Apps Script editor, delete any existing code and paste the contents of the `pkg/api/google_sheets_client.js` file
4. Update the `API_URL` constant at the top of the file to point to your deployed API:
   ```javascript
   const API_URL = "https://your-api-url.com/google-sheets";
   ```
   Replace `https://your-api-url.com` with your actual API URL.
5. Click the **Save** button (disk icon) and name your project (e.g., "Email Verifier")
6. Close the Apps Script editor and refresh your Google Sheet

### Step 3: Authorize the Script

1. After refreshing your sheet, you should see a new **Email Verifier** menu item in the menu bar
2. Click on **Email Verifier** > **Verify Selected Emails** or **Verify Column**
3. Google will prompt you to authorize the script
4. Click **Continue**
5. Select your Google account
6. You may see a warning that the app isn't verified by Google. Click **Advanced** and then **Go to Email Verifier (unsafe)**
7. Click **Allow** to grant the necessary permissions

## Using the Add-on

### Verify Selected Emails

1. Select a range of cells containing email addresses
2. Click on **Email Verifier** > **Verify Selected Emails**
3. A dialog will appear showing how many emails were found
4. Click **Yes** to confirm and start the verification process
5. The verification results will appear in adjacent columns:
   - Column to the right of the selection: Verification Status (valid, risky, or invalid)
   - Second column to the right: Confidence Score (0-100)

### Verify an Entire Column

1. Click on **Email Verifier** > **Verify Column**
2. Enter the column letter containing the emails (e.g., "A")
3. A dialog will appear showing how many emails were found
4. Click **Yes** to confirm and start the verification process
5. The verification results will appear in adjacent columns:
   - Column to the right: Verification Status (valid, risky, or invalid)
   - Second column to the right: Confidence Score (0-100)

### Verify Email Column Automatically

This feature automatically finds and processes a column with a specific name (default: "email").

1. Make sure your sheet has a column with the header "email" (or your configured column name)
2. Click on **Email Verifier** > **Verify Email Column Automatically**
3. The script will find the column and show how many emails were found
4. Click **Yes** to confirm and start the verification process
5. The verification results will appear in adjacent columns:
   - Column to the right: Verification Status (valid, risky, or invalid)
   - Second column to the right: Confidence Score (0-100)

### Configure API Settings

1. Click on **Email Verifier** > **Settings**
2. Enter your API URL or leave blank to use the default
3. Enter your preferred email column name (default is "email")
4. Click **OK** to save the settings

## Custom Functions

The integration provides custom functions that you can use directly in your spreadsheet formulas.

### EMAILVERIFY Function

The `EMAILVERIFY` function verifies a single email address and returns its verification status.

**Usage:**
```
=EMAILVERIFY(A2)
```

Where `A2` is a cell containing an email address.

**Example:**
If cell A2 contains "example@example.com", the formula `=EMAILVERIFY(A2)` might return "valid", "risky", or "invalid".

### EMAILSCORE Function

The `EMAILSCORE` function verifies a single email address and returns its confidence score.

**Usage:**
```
=EMAILSCORE(A2)
```

Where `A2` is a cell containing an email address.

**Example:**
If cell A2 contains "example@example.com", the formula `=EMAILSCORE(A2)` might return a number between 0 and 100, such as 85.

## Example Workflows

### Scenario 1: Verifying a Marketing Email List

1. Create a new Google Sheet or open an existing one with a list of email addresses
2. Ensure your Email Verifier API is deployed and running
3. Set up the Google Apps Script as described above
4. Organize your sheet with headers (e.g., "Email", "Name", "Company")
5. Click on **Email Verifier** > **Verify Email Column Automatically**
6. The script will find the "Email" column automatically
7. Confirm the operation when prompted
8. Wait for the verification process to complete
9. Once complete, you'll have two new columns:
   - "Verification Status" with values like "valid", "risky", or "invalid"
   - "Confidence Score" with numerical values from 0-100
10. Use Google Sheets' filter feature to filter out invalid emails:
    - Select all data including headers
    - Click **Data** > **Create a filter**
    - Click the filter icon on the "Verification Status" column
    - Uncheck "invalid" and click **OK**
11. You now have a cleaned email list ready for your marketing campaign

### Scenario 2: Using Custom Functions for Dynamic Verification

1. Create a new Google Sheet with a column of email addresses
2. In the column next to your emails, add the formula `=EMAILVERIFY(A2)` (assuming emails are in column A)
3. In the next column, add the formula `=EMAILSCORE(A2)`
4. Copy these formulas down to all rows with emails
5. As you add new emails to column A, the verification status and score will be automatically calculated
6. Use conditional formatting to highlight different verification statuses:
   - Select the column with verification statuses
   - Click **Format** > **Conditional formatting**
   - Add rules for each status (e.g., "valid" = green, "risky" = yellow, "invalid" = red)

## Troubleshooting

### Common Issues

#### The "Email Verifier" Menu Doesn't Appear

- Refresh the page after saving the Apps Script
- Check if there are any errors in the Apps Script editor
- Make sure you've saved the script properly

#### Verification Results Show "Error"

- Check if your API is running and accessible
- Verify that the API URL in the script settings is correct
- Check if your API server has internet access to perform email verification

#### Automatic Column Detection Not Working

- Make sure your column header exactly matches the configured email column name
- Check the settings to confirm the correct email column name
- Try using the "Verify Column" option instead and specify the column letter

#### Custom Functions Return "ERROR"

- Custom functions may take time to calculate, especially for many cells
- Check if your API is accessible from Google's servers
- Try refreshing the sheet or reloading the page
- Look for error messages in the cell (e.g., "ERROR: Invalid email format")

#### Script Runs Too Slowly

- Process emails in smaller batches
- Ensure your API server has sufficient resources
- Consider scaling up your API deployment

#### Authorization Issues

- If you see "This app isn't verified" warnings, follow the steps to proceed anyway
- If you're using the script in a corporate Google Workspace, you may need administrator approval

### Checking Logs

To view logs and debug issues:

1. Open the Apps Script editor
2. Click on **View** > **Logs**
3. Run the verification process again
4. Check the logs for any error messages

## Advanced Usage

### Customizing the Script

You can modify the Google Apps Script to suit your specific needs:

- Change the column headers for verification results
- Add additional verification information
- Implement custom formatting for different verification statuses

### Scheduling Regular Verification

You can set up triggers to automatically verify emails on a schedule:

1. In the Apps Script editor, click on **Triggers** (clock icon)
2. Click **Add Trigger**
3. Choose the function to run (e.g., `verifyEmailColumnAuto`)
4. Set the event source to **Time-driven**
5. Choose the frequency (e.g., weekly)
6. Configure the remaining settings and click **Save**

### Integrating with Other Google Workspace Apps

The verification results can be used in other Google Workspace applications:

- Create charts in Google Sheets to visualize email quality
- Use Google Forms to collect emails and verify them automatically
- Set up Google Data Studio reports based on verification results

### Using with Large Datasets

For very large email lists:

1. Split your data across multiple sheets
2. Process each sheet separately
3. Use the Google Sheets API in combination with the Email Verifier API for programmatic processing
4. Consider implementing a custom Google Sheets add-on for more advanced features 