/**
 * Email Verifier Google Sheets Add-on
 * 
 * This script provides functions to verify email addresses directly from Google Sheets
 * using the Email Verifier API.
 */

// The URL of your deployed APIf
const API_URL = "https://your-api-url.com/google-sheets";

// Default column name to look for
const DEFAULT_EMAIL_COLUMN_NAME = "email";

/**
 * Creates a custom menu in Google Sheets
 */
function onOpen() {
  const ui = SpreadsheetApp.getUi();
  ui.createMenu('Email Verifier')
    .addItem('Verify Selected Emails', 'verifySelectedEmails')
    .addItem('Verify Column', 'verifyColumn')
    .addItem('Verify Email Column Automatically', 'verifyEmailColumnAuto')
    .addItem('Settings', 'showSettings')
    .addToUi();
}

/**
 * Shows a settings dialog
 */
function showSettings() {
  const ui = SpreadsheetApp.getUi();
  const userProps = PropertiesService.getUserProperties();
  const currentApiUrl = userProps.getProperty('API_URL') || '';
  const currentEmailColumnName = userProps.getProperty('EMAIL_COLUMN_NAME') || DEFAULT_EMAIL_COLUMN_NAME;
  
  // Create a form for settings
  const htmlOutput = HtmlService.createHtmlOutput(
    '<form>' +
    '<div style="margin-bottom: 15px;">' +
    '<label for="apiUrl">API URL (leave blank to use default):</label><br>' +
    '<input type="text" id="apiUrl" name="apiUrl" style="width: 100%;" value="' + currentApiUrl + '">' +
    '</div>' +
    '<div style="margin-bottom: 15px;">' +
    '<label for="emailColumnName">Default Email Column Name:</label><br>' +
    '<input type="text" id="emailColumnName" name="emailColumnName" style="width: 100%;" value="' + currentEmailColumnName + '">' +
    '</div>' +
    '<input type="button" value="Save" onclick="google.script.host.close()" style="margin-right: 10px;">' +
    '<input type="button" value="Cancel" onclick="google.script.host.close()">' +
    '</form>'
  )
  .setWidth(400)
  .setHeight(200);
  
  const dialog = ui.showModalDialog(htmlOutput, 'Email Verifier Settings');
  
  // Get form values
  const formResponse = ui.prompt(
    'Email Verifier Settings',
    'Enter the API URL (leave blank to use default):\n\nEnter the default email column name (default: "email"):',
    ui.ButtonSet.OK_CANCEL
  );
  
  if (formResponse.getSelectedButton() == ui.Button.OK) {
    const responseText = formResponse.getResponseText();
    const parts = responseText.split('\n\n');
    
    const apiUrl = parts[0].trim();
    const emailColumnName = parts.length > 1 ? parts[1].trim() : DEFAULT_EMAIL_COLUMN_NAME;
    
    if (apiUrl) {
      userProps.setProperty('API_URL', apiUrl);
    } else {
      userProps.deleteProperty('API_URL');
    }
    
    if (emailColumnName) {
      userProps.setProperty('EMAIL_COLUMN_NAME', emailColumnName);
    } else {
      userProps.setProperty('EMAIL_COLUMN_NAME', DEFAULT_EMAIL_COLUMN_NAME);
    }
    
    ui.alert('Settings saved successfully!');
  }
}

/**
 * Gets the API URL from settings or uses the default
 */
function getApiUrl() {
  const userProps = PropertiesService.getUserProperties();
  return userProps.getProperty('API_URL') || API_URL;
}

/**
 * Gets the email column name from settings or uses the default
 */
function getEmailColumnName() {
  const userProps = PropertiesService.getUserProperties();
  return userProps.getProperty('EMAIL_COLUMN_NAME') || DEFAULT_EMAIL_COLUMN_NAME;
}

/**
 * Verifies the selected email addresses
 */
function verifySelectedEmails() {
  const ui = SpreadsheetApp.getUi();
  const sheet = SpreadsheetApp.getActiveSheet();
  const range = sheet.getActiveRange();
  
  if (!range) {
    ui.alert('Please select a range containing email addresses.');
    return;
  }
  
  const values = range.getValues();
  const emails = [];
  
  // Collect all emails from the selected range
  for (let i = 0; i < values.length; i++) {
    for (let j = 0; j < values[i].length; j++) {
      const value = values[i][j].toString().trim();
      if (value && isValidEmailFormat(value)) {
        emails.push(value);
      }
    }
  }
  
  if (emails.length === 0) {
    ui.alert('No valid email formats found in the selected range.');
    return;
  }
  
  // Confirm with the user
  const response = ui.alert(
    'Verify Emails',
    `Found ${emails.length} emails to verify. Continue?`,
    ui.ButtonSet.YES_NO
  );
  
  if (response !== ui.Button.YES) {
    return;
  }
  
  // Show loading message
  const statusRange = sheet.getRange(range.getRow(), range.getColumn() + range.getNumColumns(), 1, 2);
  statusRange.setValues([['Verification Status', 'Confidence Score']]);
  
  // Process emails in batches of 100
  const batchSize = 100;
  const results = {};
  
  for (let i = 0; i < emails.length; i += batchSize) {
    const batch = emails.slice(i, i + batchSize);
    const batchResults = verifyEmailBatch(batch);
    
    // Add results to the map
    for (const result of batchResults) {
      results[result.email.toLowerCase()] = result;
    }
  }
  
  // Update the sheet with results
  for (let i = 0; i < values.length; i++) {
    for (let j = 0; j < values[i].length; j++) {
      const email = values[i][j].toString().trim();
      if (email && isValidEmailFormat(email)) {
        const result = results[email.toLowerCase()];
        if (result) {
          // Write status and confidence score in adjacent cells
          sheet.getRange(range.getRow() + i, range.getColumn() + j + 1).setValue(result.verification_status);
          sheet.getRange(range.getRow() + i, range.getColumn() + j + 2).setValue(result.confidence_score);
        }
      }
    }
  }
  
  ui.alert(`Verification complete! Processed ${emails.length} emails.`);
}

/**
 * Verifies all emails in a specific column
 */
function verifyColumn() {
  const ui = SpreadsheetApp.getUi();
  const sheet = SpreadsheetApp.getActiveSheet();
  
  // Ask for the column
  const response = ui.prompt(
    'Verify Email Column',
    'Enter the column letter containing emails (e.g., A):',
    ui.ButtonSet.OK_CANCEL
  );
  
  if (response.getSelectedButton() !== ui.Button.OK) {
    return;
  }
  
  const columnLetter = response.getResponseText().trim().toUpperCase();
  if (!columnLetter.match(/^[A-Z]+$/)) {
    ui.alert('Please enter a valid column letter (A-Z).');
    return;
  }
  
  // Convert column letter to index
  const columnIndex = columnLetterToIndex(columnLetter);
  const dataRange = sheet.getDataRange();
  const values = dataRange.getValues();
  
  // Check if the column exists
  if (columnIndex >= values[0].length) {
    ui.alert(`Column ${columnLetter} is outside the data range.`);
    return;
  }
  
  // Collect emails from the column (starting from row 2 to skip header)
  const emails = [];
  for (let i = 1; i < values.length; i++) {
    const value = values[i][columnIndex].toString().trim();
    if (value && isValidEmailFormat(value)) {
      emails.push(value);
    }
  }
  
  if (emails.length === 0) {
    ui.alert(`No valid email formats found in column ${columnLetter}.`);
    return;
  }
  
  // Confirm with the user
  const confirmResponse = ui.alert(
    'Verify Emails',
    `Found ${emails.length} emails to verify in column ${columnLetter}. Continue?`,
    ui.ButtonSet.YES_NO
  );
  
  if (confirmResponse !== ui.Button.YES) {
    return;
  }
  
  // Add headers for results
  const statusColumnIndex = columnIndex + 1;
  const scoreColumnIndex = columnIndex + 2;
  
  // Check if we need to add new columns
  if (statusColumnIndex >= values[0].length) {
    sheet.insertColumnAfter(columnIndex);
  }
  if (scoreColumnIndex >= values[0].length + 1) { // +1 because we might have just added a column
    sheet.insertColumnAfter(statusColumnIndex);
  }
  
  // Add headers
  sheet.getRange(1, statusColumnIndex + 1).setValue('Verification Status');
  sheet.getRange(1, scoreColumnIndex + 1).setValue('Confidence Score');
  
  // Process emails in batches of 100
  const batchSize = 100;
  const results = {};
  
  for (let i = 0; i < emails.length; i += batchSize) {
    const batch = emails.slice(i, i + batchSize);
    const batchResults = verifyEmailBatch(batch);
    
    // Add results to the map
    for (const result of batchResults) {
      results[result.email.toLowerCase()] = result;
    }
  }
  
  // Update the sheet with results
  for (let i = 1; i < values.length; i++) {
    const email = values[i][columnIndex].toString().trim();
    if (email && isValidEmailFormat(email)) {
      const result = results[email.toLowerCase()];
      if (result) {
        // Write status and confidence score in adjacent columns
        sheet.getRange(i + 1, statusColumnIndex + 1).setValue(result.verification_status);
        sheet.getRange(i + 1, scoreColumnIndex + 1).setValue(result.confidence_score);
      }
    }
  }
  
  ui.alert(`Verification complete! Processed ${emails.length} emails.`);
}

/**
 * Automatically finds and verifies emails in a column with a specific name
 */
function verifyEmailColumnAuto() {
  const ui = SpreadsheetApp.getUi();
  const sheet = SpreadsheetApp.getActiveSheet();
  const dataRange = sheet.getDataRange();
  const values = dataRange.getValues();
  
  if (values.length === 0) {
    ui.alert('No data found in the sheet.');
    return;
  }
  
  // Get the configured email column name
  const emailColumnName = getEmailColumnName();
  
  // Find the email column index
  let emailColumnIndex = -1;
  const headers = values[0];
  
  for (let i = 0; i < headers.length; i++) {
    const header = headers[i].toString().trim().toLowerCase();
    if (header === emailColumnName.toLowerCase()) {
      emailColumnIndex = i;
      break;
    }
  }
  
  if (emailColumnIndex === -1) {
    ui.alert(`No column named "${emailColumnName}" found. Please check your settings or sheet headers.`);
    return;
  }
  
  // Collect emails from the column (starting from row 2 to skip header)
  const emails = [];
  for (let i = 1; i < values.length; i++) {
    const value = values[i][emailColumnIndex].toString().trim();
    if (value && isValidEmailFormat(value)) {
      emails.push(value);
    }
  }
  
  if (emails.length === 0) {
    ui.alert(`No valid email formats found in the "${emailColumnName}" column.`);
    return;
  }
  
  // Confirm with the user
  const confirmResponse = ui.alert(
    'Verify Emails',
    `Found ${emails.length} emails to verify in the "${emailColumnName}" column. Continue?`,
    ui.ButtonSet.YES_NO
  );
  
  if (confirmResponse !== ui.Button.YES) {
    return;
  }
  
  // Add headers for results
  const statusColumnIndex = emailColumnIndex + 1;
  const scoreColumnIndex = emailColumnIndex + 2;
  
  // Check if we need to add new columns
  if (statusColumnIndex >= values[0].length) {
    sheet.insertColumnAfter(emailColumnIndex);
  }
  if (scoreColumnIndex >= values[0].length + 1) { // +1 because we might have just added a column
    sheet.insertColumnAfter(statusColumnIndex);
  }
  
  // Add headers
  sheet.getRange(1, statusColumnIndex + 1).setValue('Verification Status');
  sheet.getRange(1, scoreColumnIndex + 1).setValue('Confidence Score');
  
  // Process emails in batches of 100
  const batchSize = 100;
  const results = {};
  
  for (let i = 0; i < emails.length; i += batchSize) {
    const batch = emails.slice(i, i + batchSize);
    const batchResults = verifyEmailBatch(batch);
    
    // Add results to the map
    for (const result of batchResults) {
      results[result.email.toLowerCase()] = result;
    }
  }
  
  // Update the sheet with results
  for (let i = 1; i < values.length; i++) {
    const email = values[i][emailColumnIndex].toString().trim();
    if (email && isValidEmailFormat(email)) {
      const result = results[email.toLowerCase()];
      if (result) {
        // Write status and confidence score in adjacent columns
        sheet.getRange(i + 1, statusColumnIndex + 1).setValue(result.verification_status);
        sheet.getRange(i + 1, scoreColumnIndex + 1).setValue(result.confidence_score);
      }
    }
  }
  
  ui.alert(`Verification complete! Processed ${emails.length} emails.`);
}

/**
 * Custom function to verify a single email directly in a cell
 * 
 * @param {string} email The email address to verify
 * @return {string} The verification status
 * @customfunction
 */
function EMAILVERIFY(email) {
  if (!email || typeof email !== 'string') {
    return "ERROR: Invalid input";
  }
  
  email = email.toString().trim();
  
  if (!isValidEmailFormat(email)) {
    return "ERROR: Invalid email format";
  }
  
  try {
    const results = verifyEmailBatch([email]);
    if (results && results.length > 0) {
      return results[0].verification_status;
    } else {
      return "ERROR: Verification failed";
    }
  } catch (error) {
    return "ERROR: " + error.toString();
  }
}

/**
 * Custom function to get the confidence score for an email
 * 
 * @param {string} email The email address to verify
 * @return {number} The confidence score (0-100)
 * @customfunction
 */
function EMAILSCORE(email) {
  if (!email || typeof email !== 'string') {
    return "ERROR: Invalid input";
  }
  
  email = email.toString().trim();
  
  if (!isValidEmailFormat(email)) {
    return "ERROR: Invalid email format";
  }
  
  try {
    const results = verifyEmailBatch([email]);
    if (results && results.length > 0) {
      return results[0].confidence_score;
    } else {
      return "ERROR: Verification failed";
    }
  } catch (error) {
    return "ERROR: " + error.toString();
  }
}

/**
 * Verifies a batch of emails using the API
 */
function verifyEmailBatch(emails) {
  try {
    const apiUrl = getApiUrl();
    const payload = {
      emails: emails
    };
    
    const options = {
      method: 'post',
      contentType: 'application/json',
      payload: JSON.stringify(payload),
      muteHttpExceptions: true
    };
    
    const response = UrlFetchApp.fetch(apiUrl, options);
    const responseCode = response.getResponseCode();
    
    if (responseCode === 200) {
      const responseData = JSON.parse(response.getContentText());
      return responseData.results || [];
    } else {
      Logger.log(`API error: ${response.getContentText()}`);
      SpreadsheetApp.getUi().alert(`API error: ${response.getResponseCode()}`);
      return [];
    }
  } catch (error) {
    Logger.log(`Error: ${error.toString()}`);
    SpreadsheetApp.getUi().alert(`Error: ${error.toString()}`);
    return [];
  }
}

/**
 * Checks if a string looks like a valid email format
 */
function isValidEmailFormat(email) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

/**
 * Converts a column letter to a zero-based index
 */
function columnLetterToIndex(letter) {
  let column = 0;
  const length = letter.length;
  for (let i = 0; i < length; i++) {
    column += (letter.charCodeAt(i) - 64) * Math.pow(26, length - i - 1);
  }
  return column - 1; // Convert to 0-based index
} 