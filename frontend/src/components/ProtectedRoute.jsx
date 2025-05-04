import { useAuth } from "../context/AuthContext";
import { Navigate } from "react-router-dom";

export default function ProtectedRoute({ children }) {
  const { token } = useAuth();

  if (token === null) {
    // still loading token from localStorage
    return null; // or <Loading />
  }

  return token ? children : <Navigate to="/login" replace />;
}
