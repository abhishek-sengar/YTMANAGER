// src/context/AuthContext.jsx
import { createContext, useContext, useState, useEffect } from "react";
import { jwtDecode } from "jwt-decode";

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [token, setToken] = useState(null);
  const [role, setRole] = useState(null);
  const [userName, setUserName] = useState("");  // ← new

  useEffect(() => {
    const stored = localStorage.getItem("token");
    if (stored) {
      setToken(stored);
      try {
        const decoded = jwtDecode(stored);
        setRole(decoded.role);
        setUserName(decoded.name || decoded.email); // use `name` claim or fallback to email
      } catch (err) {
        console.error("Invalid token:", err);
      }
    }
  }, []);

  const login = (newToken) => {
    localStorage.setItem("token", newToken);
    setToken(newToken);
    try {
      const decoded = jwtDecode(newToken);
      setRole(decoded.role);
      setUserName(decoded.name || decoded.email);
    } catch (err) {
      console.error("Failed to decode login token:", err);
    }
  };

  const logout = () => {
    localStorage.removeItem("token");
    setToken(null);
    setRole(null);
    setUserName("");
  };

  return (
    <AuthContext.Provider value={{ token, role, userName, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
