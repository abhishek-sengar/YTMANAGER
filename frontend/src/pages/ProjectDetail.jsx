import { useParams } from "react-router-dom";
import { useEffect, useState } from "react";
import { useAuth } from "../context/AuthContext";
import api from "../services/api";

export default function ProjectDetail() {
  const { id } = useParams();
  const { user } = useAuth(); // assuming your context gives you user info
  const [project, setProject] = useState(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState("");

  useEffect(() => {
    async function fetchProject() {
      try {
        const res = await api.get(`/projects/${id}`);
        setProject(res.data);
      } catch (err) {
        setError("Project not found or access denied.");
      } finally {
        setLoading(false);
      }
    }

    fetchProject();
  }, [id]);

  const handleAction = async (action) => {
    try {
      await api.post(`/projects/${id}/${action}`); // action = "approve" or "reject"
      setProject((prev) => ({
        ...prev,
        status: action === "approve" ? "approved" : "rejected",
      }));
      setMessage(`Project ${action}d successfully.`);
    } catch (err) {
      setMessage("Action failed. You may not have permission.");
    }
  };

  if (loading) return <p className="text-center mt-10">Loading...</p>;
  if (error) return <p className="text-center text-red-500 mt-10">{error}</p>;
  const isOwner = user?.sub === project.owner_id;

  return (
    <div className="max-w-4xl mx-auto mt-10 p-4 bg-white shadow rounded">
      <h1 className="text-2xl font-bold mb-4">{project.title}</h1>
      <p className="mb-2 text-gray-700">{project.description}</p>
      <p className="text-sm text-gray-600 mb-2">
        <strong>Status:</strong>{" "}
        <span className="capitalize font-medium">{project.status}</span>
      </p>
      <p className="text-xs text-gray-400">
        Created: {new Date(project.created_at).toLocaleString()}
      </p>
      <p className="text-xs text-gray-400 mb-4">
        Updated: {new Date(project.updated_at).toLocaleString()}
      </p>

      {project.video_path && (
        <div className="bg-gray-100 p-2 rounded mb-4 text-sm text-gray-800">
          <p className="text-sm text-gray-600 mb-1">Video Path:</p>
          {project.video_path}
        </div>
      )}

      {isOwner && project.status !== "" && (
        <div className="flex gap-4 mt-4">
          <button
            onClick={() => handleAction("approve")}
            className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
          >
            Approve
          </button>
          <button
            onClick={() => handleAction("reject")}
            className="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
          >
            Reject
          </button>
        </div>
      )}

      {message && <p className="mt-4 text-sm text-blue-600">{message}</p>}
    </div>
  );
}
