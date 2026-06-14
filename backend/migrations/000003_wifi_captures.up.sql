-- WiFi Captures table to track pentest captures
CREATE TABLE IF NOT EXISTS wifi_captures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ssid VARCHAR(255) NOT NULL,
    bssid VARCHAR(17) NOT NULL,
    channel INT,
    security VARCHAR(50),
    capture_path VARCHAR(500) NOT NULL,
    capture_name VARCHAR(255) NOT NULL,
    file_size BIGINT DEFAULT 0,
    has_handshake BOOLEAN DEFAULT FALSE,
    interface_used VARCHAR(50),
    duration_seconds INT,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'running', -- running, completed, stopped, failed
    cracked BOOLEAN DEFAULT FALSE,
    cracked_password VARCHAR(255),
    cracked_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Index for faster lookups
CREATE INDEX idx_wifi_captures_ssid ON wifi_captures(ssid);
CREATE INDEX idx_wifi_captures_bssid ON wifi_captures(bssid);
CREATE INDEX idx_wifi_captures_status ON wifi_captures(status);
CREATE INDEX idx_wifi_captures_has_handshake ON wifi_captures(has_handshake);
