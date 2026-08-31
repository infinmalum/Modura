import { useOutletContext } from "react-router-dom";

export function usePermissions() {
  return useOutletContext<{ granted: ReadonlySet<string> }>().granted;
}
