## ADDED Requirements

### Requirement: Local Watchlist Persistence
The mobile application SHALL persist the user's watchlist locally so that it remains available across app restarts.

#### Scenario: Save watchlist
- **WHEN** a user adds a new stock to the watchlist
- **THEN** the updated watchlist is saved to local storage

#### Scenario: Load watchlist on launch
- **WHEN** the user launches the application
- **THEN** the watchlist is loaded from local storage and displayed on the Home screen

### Requirement: Import Stock from Image
The mobile application SHALL allow users to import stocks by selecting an image from their device gallery.

#### Scenario: Import flow
- **WHEN** the user taps the "Import Image" button
- **THEN** the system opens the device image picker
- **WHEN** the user selects an image
- **THEN** the image is sent to the server for recognition
- **AND** the recognized stocks are presented to the user for confirmation
