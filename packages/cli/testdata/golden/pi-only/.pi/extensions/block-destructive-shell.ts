import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";

export default function blockDestructiveShell(pi: ExtensionAPI): void {
  pi.on("tool_call", async (event, ctx) => {
    if (event.toolName !== "bash") return;
    const cmd = String(event.input.command ?? "");
    if (!/\brm\s+-rf\s+\//.test(cmd)) return;
    if (ctx.hasUI) {
      const allow = await ctx.ui.confirm("Dangerous command", `This deletes from root:\n${cmd}\n\nProceed?`);
      if (allow) return;
    }
    return { block: true, reason: "rm -rf / blocked by safety policy" };
  });
}
