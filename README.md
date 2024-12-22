# Peergrine
Peergrine is an instant messaging application designed for seamless, private communication without the need for user account registration. Each time a user accesses the platform, a new, temporary identity is automatically created. Peergrine operates entirely within mainstream web browsers and requires a stable internet connection to ensure smooth real-time communication.
## Key Features

- **Instant Messaging**: Engage in private text-based conversations with selected contacts.
- **File Transfer**: Share files such as images, documents, and more with other users.
- **No Account Registration**: A new identity is created upon each session, eliminating the need for account management.
- **Cross-Device Compatibility**: Works on any major browser without requiring additional software installations.

## Development Background

Peergrine was created to address a simple yet important need: facilitating fast, secure data transfer between unfamiliar or temporary devices without the hassle of software installation. This makes Peergrine ideal for quick, on-the-go communication and file sharing.

----
## System Components

- **[JWTIssuer](./services/jwtissuer/README.md)**: Manages identity distribution through JSON Web Tokens.
- **[RtcBridge](./services/rtc-bridge/README.md)**: Handles WebRTC peer-to-peer signaling.
- **[MsgBridge](./services/msg-bridge/README.md)**: Facilitates message relaying between users.

----
## Third-Party Dependencies

Peergrine leverages distributed technologies to ensure efficient performance:
- **Centralized Services**: None
- **Distributed Services**:
	- **Pulsar**: Enables horizontal data communication between services.
	- **Redis**: Provides real-time data caching and storage.

----
## Getting Started

1. Open the Peergrine webpage in any modern web browser.
2. No account setup is required—your identity is generated automatically.
3. Enter the recipient's identifier to begin private messaging or file sharing.

For additional details on individual components or deployment, refer to the [documentation](./services).