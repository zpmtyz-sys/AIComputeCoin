import { FastifyInstance } from "fastify";
import websocket from "@fastify/websocket";
import type { WebSocket } from "ws";

interface WsClient {
  socket: WebSocket;
  channels: Set<string>;
  authenticated: boolean;
  userId?: string;
}

export async function websocketPlugin(app: FastifyInstance): Promise<void> {
  await app.register(websocket);

  const clients: Map<string, WsClient> = new Map();
  let clientIdCounter = 0;

  app.get("/ws", { websocket: true }, (socket) => {
    const ws = socket as unknown as WebSocket;
    const clientId = `client-${++clientIdCounter}`;
    const client: WsClient = {
      socket: ws,
      channels: new Set(),
      authenticated: false,
    };
    clients.set(clientId, client);

    ws.on("message", (data: Buffer) => {
      try {
        const message = JSON.parse(data.toString()) as {
          type: string;
          channel?: string;
          token?: string;
        };

        switch (message.type) {
          case "subscribe":
            if (message.channel) {
              // Private channels require authentication
              const privateChannels = ["orders", "positions", "balance"];
              if (
                privateChannels.includes(message.channel) &&
                !client.authenticated
              ) {
                ws.send(
                  JSON.stringify({
                    type: "error",
                    message: "Authentication required for private channels",
                  })
                );
                return;
              }
              client.channels.add(message.channel);
              ws.send(
                JSON.stringify({
                  type: "subscribed",
                  channel: message.channel,
                })
              );
            }
            break;

          case "unsubscribe":
            if (message.channel) {
              client.channels.delete(message.channel);
              ws.send(
                JSON.stringify({
                  type: "unsubscribed",
                  channel: message.channel,
                })
              );
            }
            break;

          case "authenticate":
            if (message.token) {
              try {
                const decoded = app.jwt.verify<{
                  userId: string;
                  email: string;
                  role: string;
                }>(message.token);
                client.authenticated = true;
                client.userId = decoded.userId;
                ws.send(
                  JSON.stringify({
                    type: "authenticated",
                    userId: decoded.userId,
                  })
                );
              } catch {
                ws.send(
                  JSON.stringify({
                    type: "error",
                    message: "Invalid authentication token",
                  })
                );
              }
            }
            break;

          default:
            ws.send(
              JSON.stringify({
                type: "error",
                message: `Unknown message type: ${message.type}`,
              })
            );
        }
      } catch {
        ws.send(
          JSON.stringify({
            type: "error",
            message: "Invalid JSON message",
          })
        );
      }
    });

    ws.on("close", () => {
      clients.delete(clientId);
    });
  });
}
