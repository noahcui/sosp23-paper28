#!/usr/bin/env python3

import argparse
import pandas as pd
import matplotlib.pyplot as plt

def convert_to_milliseconds(val):
    """Convert Go's time.duration (nanoseconds) to milliseconds"""
    return val / 1e6

def main(input_file, column, output_file):
    # Load the CSV data into a pandas DataFrame
    df = pd.read_csv(input_file)
    df.columns = df.columns.str.strip()
    # Convert the Unix timestamp to datetime and set it as the index
    df['Time'] = pd.to_datetime(df.iloc[:, 0], unit='s')

    # Subtract the first timestamp to start from 0
    df['Time'] = df['Time'] - df['Time'].iloc[0]
    df.set_index('Time', inplace=True)

    # Convert to milliseconds if the column is 'forwarding_latencies'
    if column == 'forwarding_latencies':
        print(df)
        df[column] = df[column].apply(convert_to_milliseconds)
        ylabel = 'Forward Latency (ms)'
    else:
        ylabel = column

    # Compute the average per second
    df_resampled = df[column].resample('S').mean()

    # Plot the data
    plt.plot(df_resampled.index.total_seconds(), df_resampled)
    plt.xlabel('Time (s)')
    plt.ylabel(ylabel)
    plt.title(f'Average {column} over time')

    # Save the figure to a PDF file
    plt.savefig(output_file)

if __name__ == '__main__':
    # Setup the argument parser
    parser = argparse.ArgumentParser(description='Plot data from a CSV file.')
    parser.add_argument('-input', type=str, required=True, help='Input CSV file.')
    parser.add_argument('-y', type=str, required=True, help='Column to plot.')
    parser.add_argument('-outfile', type=str, required=True, help='Output PDF file.')

    # Parse the arguments
    args = parser.parse_args()

    # Run the main function
    main(args.input, args.y, args.outfile)
