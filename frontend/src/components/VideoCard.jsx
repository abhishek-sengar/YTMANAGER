// export default function VideoCard({ video }) {
//     return (
//       <div className="border rounded shadow-sm p-4 bg-white hover:shadow-md transition">
//         <h2 className="text-lg font-semibold text-gray-800">{video.title}</h2>
//         <p className="text-sm text-gray-600 mb-1">{video.description}</p>
//         <p className="text-xs text-gray-500">
//           Channel: {video.channel_name} | Owner: {video.owner_name}
//         </p>
//         <p className="text-xs mt-2">
//           Status:{" "}
//           <span className="inline-block font-semibold px-2 py-1 rounded bg-gray-100">
//             {video.status}
//           </span>
//         </p>
//       </div>
//     );
//   }
  

// src/components/VideoCard.jsx
import { Card, CardContent, Typography, Box, Button, Stack } from "@mui/material";
import CloudUploadIcon from "@mui/icons-material/CloudUpload";
import TopBar from "../components/TopBar";

export default function VideoCard({ video, onUpload }) {
  return (
    <Card
      variant="outlined"
      sx={{
        borderRadius: 3,
        boxShadow: 2,
        bgcolor: "#fff",
        minHeight: 315,
        display: "flex",
        flexDirection: "column",
        justifyContent: "space-between",
      }}
    >
      <Box
        sx={{
          width: "100%",
          height: 180,
          bgcolor: "grey.200",
          borderTopLeftRadius: 12,
          borderTopRightRadius: 12,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <Typography variant="body2" color="text.secondary">
          4:3 Video<br />420 x 315
        </Typography>
      </Box>
      <CardContent>
        <Typography variant="subtitle1" fontWeight={600} gutterBottom noWrap>
          {video.title}
        </Typography>
        <Typography variant="body2" color="text.secondary" gutterBottom noWrap>
          {video.description}
        </Typography>
        <Typography variant="caption" color="text.secondary" display="block" gutterBottom>
          Status:{" "}
          <span style={{
            color: video.status === "approved" ? "#388e3c" : "#757575",
            fontWeight: 600
          }}>
            {video.status}
          </span>
        </Typography>
        {/* Upload to YouTube button for approved videos */}
        {video.status === "approved" && (
          <Stack direction="row" spacing={1} mt={2}>
            <Button
              variant="contained"
              color="error"
              size="small"
              startIcon={<CloudUploadIcon />}
              onClick={() => onUpload && onUpload(video)}
            >
              Upload to YouTube
            </Button>
          </Stack>
        )}
      </CardContent>
      <TopBar onMenuClick={() => {/* handle menu open/close */}} />
    </Card>
  );
}
