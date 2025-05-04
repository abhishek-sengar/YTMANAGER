import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { Button } from "../components/ui/button";
import { Card, CardContent } from "../components/ui/card";
import { Checkbox } from "../components/ui/checkbox";
import { toast } from "sonner";

export default function ChannelSelection() {
  const [channels, setChannels] = useState([]);
  const [selectedChannels, setSelectedChannels] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const { token } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const oauthChannels = localStorage.getItem("oauth_channels");
    if (oauthChannels) {
      try {
        const parsedChannels = JSON.parse(oauthChannels);
        setChannels(parsedChannels);
        // Select all channels by default
        setSelectedChannels(parsedChannels.map(ch => ch.id));
      } catch (error) {
        console.error("Error parsing channels:", error);
        toast.error("Failed to load channels");
      } finally {
        setIsLoading(false);
        // Clear the stored channels data
        localStorage.removeItem("oauth_channels");
      }
    } else {
      toast.error("No channels data found");
      navigate("/dashboard");
    }
  }, [navigate]);

  const handleChannelSelect = (channelId) => {
    setSelectedChannels((prev) =>
      prev.includes(channelId)
        ? prev.filter((id) => id !== channelId)
        : [...prev, channelId]
    );
  };

  const handleSubmit = async () => {
    if (selectedChannels.length === 0) {
      toast.error("Please select at least one channel");
      return;
    }

    try {
      const selectedChannelData = channels.filter((ch) =>
        selectedChannels.includes(ch.id)
      );

      const response = await fetch("/api/youtube/add-channels", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ channels: selectedChannelData }),
      });

      if (!response.ok) {
        throw new Error("Failed to add channels");
      }

      toast.success("Channels added successfully");
      navigate("/dashboard");
    } catch (error) {
      console.error("Error adding channels:", error);
      toast.error("Failed to add channels");
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-8">Select Your YouTube Channels</h1>
        
        {channels.length === 0 ? (
          <Card className="w-full">
            <CardContent className="p-6">
              <p className="text-center text-muted-foreground">
                No channels found. Please try connecting your YouTube account again.
              </p>
            </CardContent>
          </Card>
        ) : (
          <div className="space-y-4">
            {channels.map((channel) => (
              <Card key={channel.id} className="w-full">
                <CardContent className="p-4 flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <img
                      src={channel.iconUrl}
                      alt={channel.name}
                      className="w-12 h-12 rounded-full"
                    />
                    <div>
                      <h3 className="font-medium">{channel.name}</h3>
                    </div>
                  </div>
                  <Checkbox
                    checked={selectedChannels.includes(channel.id)}
                    onCheckedChange={(checked) => {
                      if (checked) {
                        handleChannelSelect(channel.id);
                      } else {
                        handleChannelSelect(channel.id);
                      }
                    }}
                    className="h-4 w-4"
                  />
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        <div className="mt-8 flex justify-end">
          <Button
            onClick={handleSubmit}
            disabled={selectedChannels.length === 0}
            className="bg-primary text-primary-foreground hover:bg-primary/90"
          >
            Add Selected Channels
          </Button>
        </div>
      </div>
    </div>
  );
} 