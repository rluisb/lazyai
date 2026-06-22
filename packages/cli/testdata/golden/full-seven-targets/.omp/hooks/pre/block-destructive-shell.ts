import type { HookAPI } from "@oh-my-pi/pi-coding-agent/extensibility/hooks";

export default function blockDestructiveShell(omp: HookAPI): void {
  omp.on("tool_call", async (event, ctx) => {
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
