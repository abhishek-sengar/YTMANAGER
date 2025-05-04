// import VideoCard from "./VideoCard";

// export default function ProjectGrid({ videos }) {
//   return (
//     <div>
//       <h2 className="text-lg font-semibold mb-4 text-gray-700">Recent Projects</h2>
//       {!videos || videos.length === 0 ? (
//         <p className="text-gray-500 text-sm">No projects found for this selection.</p>
//       ) : (
//         <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
//           {videos.map((video) => (
//             <VideoCard key={video.id} video={video} />
//           ))}
//         </div>
//       )}
//     </div>
//   );
// }


// src/components/ProjectGroup.jsx
import { Typography, Grid, Divider, Box, Avatar } from "@mui/material";
import VideoCard from "./VideoCard";

export default function ProjectGroup({ group, videos, onUpload }) {
  return (
    <Box mb={4}>
      <Box display="flex" alignItems="center" mb={1}>
        <Avatar sx={{ mr: 1 }}>{group?.[0]?.toUpperCase() || "?"}</Avatar>
        <Typography variant="h6" fontWeight={600}>
          {group}
        </Typography>
      </Box>
      <Divider sx={{ mb: 2 }} />
      <Grid container spacing={3}>
        {videos.map((video) => (
          <Grid item xs={12} sm={6} md={4} lg={3} key={video.id}>
            <VideoCard video={video} onUpload={onUpload} />
          </Grid>
        ))}
      </Grid>
    </Box>
  );
}
