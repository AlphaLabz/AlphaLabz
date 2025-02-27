import { useLocation } from "react-router-dom"

export default function usePathname() {
  const location = useLocation()
  const pathname = location.pathname
  const segments = location.pathname.split("/").filter(Boolean)

  const paths = segments.map((segment, index) => ({
    name: segment,
    url: "/" + segments.slice(0, index + 1).join("/"), // Build full URL
  }))
  return { pathname, paths }
}
